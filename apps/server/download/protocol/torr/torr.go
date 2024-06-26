package torr

import (
	"downite/db"
	"downite/types"
	"downite/utils"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	gotorrenttypes "github.com/anacrolix/torrent/types"
	"github.com/anacrolix/torrent/types/infohash"
	"modernc.org/sqlite"
)

type TorrentPrevSize struct {
	DownloadedBytes int64
	UploadedBytes   int64
}

type TorrentClient struct {
	Client             *gotorrent.Client
	torrentPrevSizeMap map[string]TorrentPrevSize
	TorrentQueue       []string
	mutexForTorrents   sync.Mutex
	Torrents           map[string]*types.Torrent
	Config             *types.TorrentClientConfig
}

func CreateTorrentClient(config types.TorrentClientConfig) (*TorrentClient, error) {
	torrentClient := &TorrentClient{
		Config:             &config,
		torrentPrevSizeMap: make(map[string]TorrentPrevSize),
		TorrentQueue:       make([]string, 0),
	}
	// Create a new torrent client config
	goTorrentClientConfig := gotorrent.NewDefaultClientConfig()
	sqliteStorage, err := storage.NewSqlitePieceCompletion(config.PieceCompletionDbPath)
	if err != nil {
		fmt.Printf("Error creating sqlite storage: %v\n", err)
		return nil, err
	}
	goTorrentClientConfig.DefaultStorage = storage.NewFileWithCompletion(config.DownloadPath, sqliteStorage)

	// Initialize the gotorrent client
	client, err := gotorrent.NewClient(goTorrentClientConfig)
	if err != nil {
		fmt.Println("Error creating gotorrent client:", err)
		return nil, err
	}
	torrentClient.Client = client
	return torrentClient, nil
}
func (torrentClient *TorrentClient) InitTorrents() error {
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

		torrentClient.mutexForTorrents.Lock()
		torrentClient.Torrents[dbTorrent.Infohash] = &dbTorrent
		torrentClient.mutexForTorrents.Unlock()

		go func() {
			torrent, err := torrentClient.AddTorrent(dbTorrent.Infohash, dbTorrent.Trackers, dbTorrent.SavePath, true)
			if err != nil {
				fmt.Printf("Error while adding torrent to client %s", err)
			}
			if dbTorrent.Status == types.TorrentStatusDownloading.String() {
				_, err = torrentClient.StartTorrent(torrent)
				if err != nil {
					fmt.Printf("Error while starting torrent download %s", err)
				}
			}
		}()
	}
	// Start a goroutine to update download speed
	go torrentClient.updateTorrentSpeeds()
	// Start a goroutine to check completed torrents
	go torrentClient.checkCompletedTorrents()
	return nil
}
func (torrentClient *TorrentClient) checkCompletedTorrents() {
	for {
		torrents := torrentClient.Client.Torrents()
		for _, torrent := range torrents {
			torrentClient.mutexForTorrents.Lock()
			dbTorrent := torrentClient.Torrents[torrent.InfoHash().String()]
			if dbTorrent.Status != types.TorrentStatusDownloading.String() {
				continue
			}

			done := false
			for _, file := range torrent.Files() {
				if file.Priority() == gotorrenttypes.PiecePriorityNone {
					done = true
					continue
				}
				if file.BytesCompleted() == file.Length() {
					done = true
				}
			}
			if done {
				db.UpdateTorrentStatus(torrent.InfoHash().String(), types.TorrentStatusCompleted)
				dbTorrent.Status = types.TorrentStatusCompleted.String()
			}
			torrentClient.mutexForTorrents.Unlock()
		}
		time.Sleep(time.Second / 2)
	}
}
func (torrentClient *TorrentClient) updateTorrentSpeeds() {
	for {
		torrents := torrentClient.Client.Torrents()
		for _, torrent := range torrents {
			prevDownloadedTotalLength := torrentClient.torrentPrevSizeMap[torrent.InfoHash().HexString()].DownloadedBytes
			newDownloadedTotalLength := torrent.BytesCompleted()
			downloadedByteCount := newDownloadedTotalLength - prevDownloadedTotalLength
			downloadSpeed := float32(downloadedByteCount) / 1024

			prevUploadedTotalLength := torrentClient.torrentPrevSizeMap[torrent.InfoHash().HexString()].UploadedBytes
			stats := torrent.Stats()
			uploadedByteCount := stats.BytesWrittenData.Int64() - prevUploadedTotalLength
			uploadSpeed := float32(uploadedByteCount) / 1024

			prevSize := torrentClient.torrentPrevSizeMap[torrent.InfoHash().HexString()]
			prevSize.DownloadedBytes = newDownloadedTotalLength
			prevSize.UploadedBytes = stats.BytesWrittenData.Int64()
			torrentClient.torrentPrevSizeMap[torrent.InfoHash().HexString()] = prevSize

			// set torrent speed info
			torrentClient.mutexForTorrents.Lock()
			dbTorrent := torrentClient.Torrents[torrent.InfoHash().HexString()]
			dbTorrent.DownloadSpeed = downloadSpeed
			dbTorrent.UploadSpeed = uploadSpeed
			torrentClient.mutexForTorrents.Unlock()
		}
		time.Sleep(time.Second)
	}

}
func (torrentClient *TorrentClient) RegisterTorrent(infohash string,
	name string,
	savePath string,
	specTrackers [][]string) (*types.Torrent, error) {

	var err error

	// if save path empty use default path
	if savePath == "" {
		savePath = torrentClient.Config.DownloadPath
	} else {
		if err = utils.CheckDirectoryExists(savePath); err != nil {
			return nil, err
		}
	}

	dbTorrent := types.Torrent{
		Infohash: infohash,
		Name:     name,
		SavePath: savePath,
		Status:   types.TorrentStatusStringMap[types.TorrentStatusMetadata],
		Trackers: []types.Tracker{},
	}

	for tierIndex, trackersOfTier := range specTrackers {
		for _, tracker := range trackersOfTier {
			//validate url
			trackerUrl, err := url.Parse(tracker)
			if err != nil {
				return nil, err
			}
			dbTorrent.Trackers = append(dbTorrent.Trackers, types.Tracker{
				Url:  trackerUrl.String(),
				Tier: tierIndex,
			})
		}
	}

	// Insert torrent
	err = db.InsertTorrent(&dbTorrent)
	if err != nil {
		return nil, err
	}

	// Insert trackers
	for _, dbTracker := range dbTorrent.Trackers {
		if err = db.InsertTracker(&dbTracker, dbTorrent.Infohash); err != nil {
			// if error is not 2067 (duplicate key) then return
			if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() != 2067 {
				return nil, err
			}
		}
	}

	torrentClient.mutexForTorrents.Lock()
	torrentClient.Torrents[dbTorrent.Infohash] = &dbTorrent
	torrentClient.mutexForTorrents.Unlock()

	return &dbTorrent, nil
}

func (torrentClient *TorrentClient) AddTorrent(hash string, trackers []types.Tracker, savePath string, verifyFiles bool) (*gotorrent.Torrent, error) {
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
	torrent, new, err := torrentClient.Client.AddTorrentSpec(&torrentSpec)
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
func (torrentClient *TorrentClient) RegisterFiles(infohash metainfo.Hash, inputFiles *[]types.TorrentFileFlatTreeNode) (*types.Torrent, error) {
	torrent, ok := torrentClient.Client.Torrent(infohash)
	if !ok {
		return nil, fmt.Errorf("cannot find torrent with %s this infohash", infohash)
	}

	torrentClient.mutexForTorrents.Lock()
	dbTorrent := torrentClient.Torrents[infohash.String()]
	// Insert download priorities of the files
	for _, file := range torrent.Files() {
		for _, clientFile := range *inputFiles {
			if file.DisplayPath() == clientFile.Path {
				priority, ok := types.PiecePriorityStringMap[clientFile.Priority]
				if !ok {
					return nil, fmt.Errorf("invalid download priority: %s", clientFile.Priority)
				}

				if priority != gotorrenttypes.PiecePriorityNone {
					dbTorrent.SizeOfWanted += file.Length()
				}

				var fileName string
				//if its not multi file torrentt path array gonna be empty. use display path instead
				if len(file.FileInfo().Path) == 0 {
					fileName = file.DisplayPath()
				} else {
					fileName = file.FileInfo().Path[len(file.FileInfo().Path)-1]
				}
				db.InsertTorrentFile(&types.TorrentFileTreeNode{
					Path:     file.Path(),
					Priority: clientFile.Priority,
					Name:     fileName,
				}, dbTorrent.Infohash)
			}
		}

	}
	db.UpdateSizeOfWanted(dbTorrent)
	dbTorrent.Files = CreateFileTreeFromMeta(*torrent.Info())
	torrentClient.mutexForTorrents.Unlock()

	return dbTorrent, nil
}
func (torrentClient *TorrentClient) StartTorrent(torrent *gotorrent.Torrent) (*gotorrent.Torrent, error) {
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
	torrentClient.torrentPrevSizeMap[torrent.InfoHash().String()] = TorrentPrevSize{
		DownloadedBytes: torrent.BytesCompleted(),
		UploadedBytes:   0,
	}

	return torrent, nil
}
func (torrentClient *TorrentClient) FindTorrents(hashes []string) ([]*gotorrent.Torrent, error) {
	foundTorrents := []*gotorrent.Torrent{}
	for _, hash := range hashes {
		torrent, ok := torrentClient.Client.Torrent(infohash.FromHexString(hash))
		if !ok {
			return nil, fmt.Errorf("torrent with hash %s not found", hash)
		}

		foundTorrents = append(foundTorrents, torrent)
	}
	return foundTorrents, nil
}

//	func GetTorrentDetails(torrent *gotorrent.Torrent) (*types.Torrent, error) {
//		var foundTorrent types.Torrent
//		dbTorrent, err := db.GetTorrent(torrent.InfoHash().String())
//		if err != nil {
//			return nil, err
//		}
//		//info is not yet received
//		if torrent.Info() == nil {
//			foundTorrent = types.Torrent{
//				Infohash:  torrent.InfoHash().String(),
//				Name:      torrent.Name(),
//				CreatedAt: dbTorrent.CreatedAt,
//				Status:    types.TorrentStatusStringMap[types.TorrentStatusMetadata],
//			}
//		} else {
//			fileTree := CreateFileTreeFromMeta(*torrent.Info())
//			var progress float32 = 0.0
//			if dbTorrent.SizeOfWanted != 0 {
//				progress = float32(torrent.BytesCompleted()) / float32(dbTorrent.SizeOfWanted) * 100
//			}
//			trackers := []types.Tracker{}
//
//			torrentMeta := torrent.Metainfo()
//			torrentSpec := gotorrent.TorrentSpecFromMetaInfo(&torrentMeta)
//			specTrackers := torrentSpec.Trackers
//
//			for tierIndex, trackersOfTier := range specTrackers {
//				for _, tracker := range trackersOfTier {
//					trackers = append(trackers, types.Tracker{
//						Url:   tracker,
//						Tier:  tierIndex,
//						Peers: []types.Peer{},
//					})
//				}
//			}
//
//			torrentPeers := torrent.PeerConns()
//			peers := []types.Peer{}
//			for _, peer := range torrentPeers {
//				peers = append(peers, types.Peer{
//					Url: peer.RemoteAddr.String(),
//				})
//			}
//			foundTorrent = types.Torrent{
//				Infohash:     torrent.InfoHash().String(),
//				Name:         torrent.Name(),
//				QueueNumber:  dbTorrent.QueueNumber,
//				CreatedAt:    dbTorrent.CreatedAt,
//				Files:        fileTree,
//				TotalSize:    torrent.Info().TotalLength(),
//				SizeOfWanted: dbTorrent.SizeOfWanted,
//				AmountLeft:   torrent.BytesMissing(),
//				Downloaded:   torrent.BytesCompleted(),
//				Progress:     progress,
//				Seeds:        torrent.Stats().ConnectedSeeders,
//				PeerCount:    torrent.Stats().ActivePeers,
//				Status:       dbTorrent.Status,
//				Trackers:     trackers,
//				Peers:        peers,
//			}
//
//
//		}
//		return &foundTorrent, nil
//	}
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

func (torrentClient *TorrentClient) GetTotalDownloadSpeed() float32 {
	torrentClient.mutexForTorrents.Lock()
	defer torrentClient.mutexForTorrents.Unlock()
	var totalDownloadSpeed float32
	for _, torrent := range torrentClient.Torrents {
		totalDownloadSpeed += torrent.DownloadSpeed
	}
	return totalDownloadSpeed
}
func (torrentClient *TorrentClient) GetTotalUploadSpeed() float32 {
	torrentClient.mutexForTorrents.Lock()
	defer torrentClient.mutexForTorrents.Unlock()
	var totalUploadSpeed float32
	for _, torrent := range torrentClient.Torrents {
		totalUploadSpeed += torrent.UploadSpeed
	}
	return totalUploadSpeed
}
