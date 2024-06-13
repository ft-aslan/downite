package torr

import (
	"downite/db"
	"downite/types"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
)

var Client *gotorrent.Client

type TorrentPrevSize struct {
	DownloadedBytes int64
	UploadedBytes   int64
}

var (
	MutexForTorrentSpeed sync.Mutex
	TorrentSpeedMap      = make(map[string]types.TorrentSpeedInfo)
	torrentPrevSizeMap   = make(map[string]TorrentPrevSize)
	TorrentQueue         = make([]string, 0)
	TorrentClientConfig  *types.TorrentClientConfig
)

func CreateTorrentClient(config types.TorrentClientConfig) error {
	TorrentClientConfig = &config
	// Create a new torrent client config
	goTorrentClientConfig := gotorrent.NewDefaultClientConfig()
	sqliteStorage, err := storage.NewSqlitePieceCompletion(config.PieceCompletionDbPath)
	if err != nil {
		fmt.Printf("Error creating sqlite storage: %v\n", err)
		return err
	}
	goTorrentClientConfig.DefaultStorage = storage.NewFileWithCompletion(config.DownloadPath, sqliteStorage)

	// Initialize the torrent client
	Client, err = gotorrent.NewClient(goTorrentClientConfig)
	if err != nil {
		fmt.Println("Error creating torrent client:", err)
		return err
	}
	return nil
}
func InitTorrents() error {
	dbTorrents, err := db.GetTorrents()
	if err != nil {
		return err
	}
	for _, dbTorrent := range dbTorrents {
		// get the trackers
		trackers, err := db.GetTorrentTrackers(dbTorrent.Infohash)
		if err != nil {
			return err
		}
		dbTorrent.Trackers = trackers

		go func() {
			torrent, err := AddTorrent(dbTorrent.Infohash, dbTorrent.Trackers, dbTorrent.SavePath, true)
			if err != nil {
				fmt.Printf("Error while adding torrent to client %s", err)
			}
			if dbTorrent.Status == types.TorrentStatusDownloading.String() {
				_, err = StartTorrent(torrent)
				if err != nil {
					fmt.Printf("Error while starting torrent download %s", err)
				}
			}
		}()
	}
	// Start a goroutine to update download speed
	go updateTorrentSpeeds()

	return nil
}
func updateTorrentSpeeds() {
	for {
		torrents := Client.Torrents()
		for _, torrent := range torrents {
			MutexForTorrentSpeed.Lock()

			prevDownloadedTotalLength := torrentPrevSizeMap[torrent.InfoHash().HexString()].DownloadedBytes
			newDownloadedTotalLength := torrent.BytesCompleted()
			downloadedByteCount := newDownloadedTotalLength - prevDownloadedTotalLength
			downloadSpeed := float32(downloadedByteCount) / 1024

			prevUploadedTotalLength := torrentPrevSizeMap[torrent.InfoHash().HexString()].UploadedBytes
			stats := torrent.Stats()
			uploadedByteCount := stats.BytesWrittenData.Int64() - prevUploadedTotalLength
			uploadSpeed := float32(uploadedByteCount) / 1024

			TorrentSpeedMap[torrent.InfoHash().HexString()] = types.TorrentSpeedInfo{
				DownloadSpeed: downloadSpeed,
				UploadSpeed:   uploadSpeed,
			}
			prevSize := torrentPrevSizeMap[torrent.InfoHash().HexString()]
			prevSize.DownloadedBytes = newDownloadedTotalLength
			prevSize.UploadedBytes = stats.BytesWrittenData.Int64()
			torrentPrevSizeMap[torrent.InfoHash().HexString()] = prevSize

			MutexForTorrentSpeed.Unlock()
		}
		time.Sleep(time.Second)
	}

}

func AddTorrent(hash string, trackers []types.Tracker, savePath string, verifyFiles bool) (*gotorrent.Torrent, error) {
	torrentSpec := gotorrent.TorrentSpec{
		InfoHash: infohash.FromHexString(hash),
	}
	pieceCompletion, err := storage.NewDefaultPieceCompletionForDir("./tmp")
	if err != nil {
		return nil, fmt.Errorf("new piece completion: %w", err)
	}
	torrentSpec.Storage = storage.NewFileOpts(storage.NewFileClientOpts{
		ClientBaseDir: savePath,
		TorrentDirMaker: func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
			return filepath.Join(baseDir, info.BestName())
		},
		FilePathMaker: func(opts storage.FilePathMakerOpts) string {
			return filepath.Join(opts.File.BestPath()...)
		},
		PieceCompletion: pieceCompletion,
	})
	torrent, new, err := Client.AddTorrentSpec(&torrentSpec)
	if err != nil {
		return nil, err
	}
	if !new {
		return nil, fmt.Errorf("torrent with hash %s already exists", hash)
	}
	// set trackers of torrent
	dbTrackers := trackers
	if len(dbTrackers) > 0 {
		// sort it based on their tiers
		sort.Slice(dbTrackers, func(i, j int) bool { return dbTrackers[i].Tier < dbTrackers[j].Tier })
		// get the maximum tier number and create a tieredTrackers slice
		maximumTierIndex := dbTrackers[len(dbTrackers)-1].Tier
		tieredTrackers := make([][]string, 0, maximumTierIndex)
		// initialize the tieredTrackers slice
		for i := 0; i < maximumTierIndex+1; i++ {
			tieredTrackers = append(tieredTrackers, []string{})
		}
		// insert the trackers into the tieredTrackers slice based on their tiers
		for _, tracker := range dbTrackers {
			tieredTrackers[tracker.Tier] = append(tieredTrackers[tracker.Tier], tracker.Url)
		}
		// Add trackers to the torrent
		torrent.AddTrackers(tieredTrackers)
	}

	// we need metainfo so we wait for it
	<-torrent.GotInfo()

	// verify the torrent
	if verifyFiles {
		torrent.VerifyData()
	}

	return torrent, nil
}

func StartTorrent(torrent *gotorrent.Torrent) (*gotorrent.Torrent, error) {
	// set torrent file priorities
	// TODO(fatih): in the future we can make this a hashmap for faster search
	dbFiles, err := db.GetTorrentTorrentFiles(torrent.InfoHash().String())
	if err != nil {
		return nil, err
	}

	for _, file := range torrent.Files() {
		for _, dbFile := range dbFiles {
			if file.Path() == dbFile.Path {
				// set priority also starts the download for file if priority is not none
				file.SetPriority(types.PiecePriorityStringMap[dbFile.Priority])
			}
		}
	}

	// get current size of torrent for speed calculation
	torrentPrevSizeMap[torrent.InfoHash().String()] = TorrentPrevSize{
		DownloadedBytes: torrent.BytesCompleted(),
		UploadedBytes:   0,
	}

	return torrent, nil
}
func FindTorrents(hashes []string) ([]*gotorrent.Torrent, error) {
	foundTorrents := []*gotorrent.Torrent{}
	for _, hash := range hashes {
		torrent, ok := Client.Torrent(infohash.FromHexString(hash))
		if !ok {
			return nil, fmt.Errorf("torrent with hash %s not found", hash)
		}

		foundTorrents = append(foundTorrents, torrent)
	}
	return foundTorrents, nil
}
func GetTorrentDetails(torrent *gotorrent.Torrent) (*types.Torrent, error) {
	var foundTorrent types.Torrent
	dbTorrent, err := db.GetTorrent(torrent.InfoHash().String())
	if err != nil {
		return nil, err
	}
	//info is not yet received
	if torrent.Info() == nil {
		foundTorrent = types.Torrent{
			Infohash:  torrent.InfoHash().String(),
			Name:      torrent.Name(),
			CreatedAt: dbTorrent.CreatedAt,
			Status:    types.TorrentStatusStringMap[types.TorrentStatusMetadata],
		}
	} else {
		fileTree := CreateFileTreeFromMeta(*torrent.Info())
		var progress float32 = 0.0
		if dbTorrent.SizeOfWanted != 0 {
			progress = float32(torrent.BytesCompleted()) / float32(dbTorrent.SizeOfWanted) * 100
		}
		trackers := []types.Tracker{}

		torrentMeta := torrent.Metainfo()
		torrentSpec := gotorrent.TorrentSpecFromMetaInfo(&torrentMeta)
		specTrackers := torrentSpec.Trackers

		for tierIndex, trackersOfTier := range specTrackers {
			for _, tracker := range trackersOfTier {
				trackers = append(trackers, types.Tracker{
					Url:   tracker,
					Tier:  tierIndex,
					Peers: []types.Peer{},
				})
			}
		}

		torrentPeers := torrent.PeerConns()
		peers := []types.Peer{}
		for _, peer := range torrentPeers {
			peers = append(peers, types.Peer{
				Url: peer.RemoteAddr.String(),
			})
		}
		foundTorrent = types.Torrent{
			Infohash:     torrent.InfoHash().String(),
			Name:         torrent.Name(),
			QueueNumber:  dbTorrent.QueueNumber,
			CreatedAt:    dbTorrent.CreatedAt,
			Files:        fileTree,
			TotalSize:    torrent.Info().TotalLength(),
			SizeOfWanted: dbTorrent.SizeOfWanted,
			AmountLeft:   torrent.BytesMissing(),
			Downloaded:   torrent.BytesCompleted(),
			Progress:     progress,
			Seeds:        torrent.Stats().ConnectedSeeders,
			PeerCount:    torrent.Stats().ActivePeers,
			Status:       dbTorrent.Status,
			Trackers:     trackers,
			Peers:        peers,
		}

		// we use mutex becouse calculating speed is concurrent
		MutexForTorrentSpeed.Lock()
		speeds := TorrentSpeedMap[torrent.InfoHash().String()]
		foundTorrent.DownloadSpeed = speeds.DownloadSpeed
		foundTorrent.UploadSpeed = speeds.UploadSpeed
		MutexForTorrentSpeed.Unlock()

	}
	return &foundTorrent, nil
}
func createFolder(fileTree *[]*types.TorrentFileTreeNode, path []string) (*[]*types.TorrentFileTreeNode, *types.TorrentFileTreeNode) {
	currentFileTree := fileTree
	var parentNode *types.TorrentFileTreeNode
	for pathIndex, segment := range path {
		currentPath := path[:pathIndex+1]
		found := false
		if len(*currentFileTree) > 0 {
			for _, node := range *currentFileTree {
				if node.Name == segment {
					parentNode = node
					currentFileTree = node.Children
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		parentNode = &types.TorrentFileTreeNode{
			Length:   0,
			Name:     segment,
			Path:     strings.Join(currentPath, "/"),
			Children: &[]*types.TorrentFileTreeNode{},
		}
		*currentFileTree = append(*currentFileTree, parentNode)
		currentFileTree = parentNode.Children
	}

	return currentFileTree, parentNode
}
func CreateFileTreeFromMeta(meta metainfo.Info) []*types.TorrentFileTreeNode {
	var fileTree []*types.TorrentFileTreeNode
	//there is no file tree in torrent
	if len(meta.Files) == 0 {
		fileTree = []*types.TorrentFileTreeNode{
			{
				Length:   meta.TotalLength(),
				Name:     meta.Name,
				Path:     meta.Name,
				Children: &[]*types.TorrentFileTreeNode{},
			},
		}
	}
	//there is a file tree in torrent
	for _, file := range meta.Files {
		targetNodeTree := &fileTree
		var parentNode *types.TorrentFileTreeNode
		if len(file.Path) > 1 {
			targetNodeTree, parentNode = createFolder(targetNodeTree, file.Path[:len(file.Path)-1])
		}
		*targetNodeTree = append(*targetNodeTree, &types.TorrentFileTreeNode{
			Length:   file.Length,
			Name:     file.Path[len(file.Path)-1],
			Path:     strings.Join(file.Path, "/"),
			Children: &[]*types.TorrentFileTreeNode{},
		})
		if parentNode != nil {
			parentNode.Length += file.Length
		}
	}
	return fileTree
}

func GetTotalDownloadSpeed() float32 {
	MutexForTorrentSpeed.Lock()
	defer MutexForTorrentSpeed.Unlock()
	var totalDownloadSpeed float32
	for _, downloadSpeed := range TorrentSpeedMap {
		totalDownloadSpeed += downloadSpeed.DownloadSpeed
	}
	return totalDownloadSpeed
}
func GetTotalUploadSpeed() float32 {
	MutexForTorrentSpeed.Lock()
	defer MutexForTorrentSpeed.Unlock()
	var totalUploadSpeed float32
	for _, uploadSpeed := range TorrentSpeedMap {
		totalUploadSpeed += uploadSpeed.UploadSpeed
	}
	return totalUploadSpeed
}
