package torr

import (
	"downite/db"
	"downite/types"
	"downite/utils"
	"fmt"
	"net/url"
	"os"
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
type TorrentEngineConfig struct {
	PieceCompletionDbPath string
	DownloadPath          string
}

type TorrentEngine struct {
	client             *gotorrent.Client
	torrentPrevSizeMap map[string]TorrentPrevSize
	TorrentQueue       []string
	mutexForTorrents   sync.Mutex
	torrents           map[string]*types.Torrent
	Config             *TorrentEngineConfig
	db                 *db.Database
}

func CreateTorrentEngine(config TorrentEngineConfig, db *db.Database) (*TorrentEngine, error) {
	torrentEngine := &TorrentEngine{
		Config:             &config,
		torrentPrevSizeMap: make(map[string]TorrentPrevSize),
		TorrentQueue:       make([]string, 0),
		torrents:           make(map[string]*types.Torrent),
		db:                 db,
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
	torrentEngine.client = client
	return torrentEngine, nil
}
func (torrentEngine *TorrentEngine) InitTorrents() error {
	dbTorrents, err := torrentEngine.db.GetTorrents()
	if err != nil {
		return err
	}
	for _, dbTorrent := range dbTorrents {
		// get the trackers
		trackers, err := torrentEngine.db.GetTorrentTrackers(dbTorrent.Infohash)
		if err != nil {
			return err
		}
		dbTorrent.Trackers = trackers

		torrentEngine.mutexForTorrents.Lock()
		torrentEngine.torrents[dbTorrent.Infohash] = &dbTorrent
		torrentEngine.mutexForTorrents.Unlock()

		go func() {
			torrent, err := torrentEngine.AddTorrent(dbTorrent.Infohash, dbTorrent.Trackers, dbTorrent.SavePath, true)
			if err != nil {
				fmt.Printf("Error while adding torrent to client %s", err)
			}
			if dbTorrent.Status == types.TorrentStatusDownloading.String() {
				_, err = torrentEngine.StartTorrent(torrent)
				if err != nil {
					fmt.Printf("Error while starting torrent download %s", err)
				}
			}
		}()
	}
	// Start a goroutine to update download speed
	go torrentEngine.updateTorrentSpeeds()
	// Start a goroutine to check completed torrents
	go torrentEngine.checkCompletedTorrents()
	// Start a goroutine to update torrent info
	go torrentEngine.updateTorrentInfo()
	return nil
}
func (torrentEngine *TorrentEngine) checkCompletedTorrents() {
	for {
		torrents := torrentEngine.client.Torrents()
		torrentEngine.mutexForTorrents.Lock()
		for _, torrent := range torrents {
			dbTorrent := torrentEngine.torrents[torrent.InfoHash().String()]
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
				torrentEngine.db.UpdateTorrentStatus(torrent.InfoHash().String(), types.TorrentStatusCompleted)
				dbTorrent.Status = types.TorrentStatusCompleted.String()
			}
		}
		torrentEngine.mutexForTorrents.Unlock()
		time.Sleep(time.Second / 2)
	}
}
func (torrentEngine *TorrentEngine) updateTorrentInfo() {
	for {
		torrents := torrentEngine.client.Torrents()
		torrentEngine.mutexForTorrents.Lock()
		for _, torrent := range torrents {
			dbTorrent := torrentEngine.torrents[torrent.InfoHash().String()]

			//Update peers
			torrentPeers := torrent.PeerConns()
			peers := []types.Peer{}
			for _, peer := range torrentPeers {
				peers = append(peers, types.Peer{
					Url: peer.RemoteAddr.String(),
				})
			}
			dbTorrent.Peers = peers

			//Update progress
			var progress float32 = 0.0
			if dbTorrent.SizeOfWanted != 0 {
				progress = float32(torrent.BytesCompleted()) / float32(dbTorrent.SizeOfWanted) * 100
			}
			dbTorrent.Progress = progress

		}
		torrentEngine.mutexForTorrents.Unlock()
		time.Sleep(time.Second)
	}
}

func (torrentEngine *TorrentEngine) updateTorrentSpeeds() {
	for {
		torrents := torrentEngine.client.Torrents()
		for _, torrent := range torrents {
			// calculate torrent speed based on written bytes per sec
			prevDownloadedTotalLength := torrentEngine.torrentPrevSizeMap[torrent.InfoHash().HexString()].DownloadedBytes
			newDownloadedTotalLength := torrent.BytesCompleted()
			downloadedByteCount := newDownloadedTotalLength - prevDownloadedTotalLength
			downloadSpeed := float32(downloadedByteCount) / 1024

			prevUploadedTotalLength := torrentEngine.torrentPrevSizeMap[torrent.InfoHash().HexString()].UploadedBytes
			stats := torrent.Stats()
			uploadedByteCount := stats.BytesWrittenData.Int64() - prevUploadedTotalLength
			uploadSpeed := float32(uploadedByteCount) / 1024

			prevSize := torrentEngine.torrentPrevSizeMap[torrent.InfoHash().HexString()]
			prevSize.DownloadedBytes = newDownloadedTotalLength
			prevSize.UploadedBytes = stats.BytesWrittenData.Int64()
			torrentEngine.torrentPrevSizeMap[torrent.InfoHash().HexString()] = prevSize

			// set torrent speed info
			torrentEngine.mutexForTorrents.Lock()
			dbTorrent := torrentEngine.torrents[torrent.InfoHash().HexString()]
			dbTorrent.DownloadSpeed = downloadSpeed
			dbTorrent.UploadSpeed = uploadSpeed
			torrentEngine.mutexForTorrents.Unlock()
		}
		time.Sleep(time.Second)
	}

}
func (torrentEngine *TorrentEngine) GetTorrents() []*types.Torrent {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()
	return torrentEngine.torrentsAsSlice()
}
func (torrentEngine *TorrentEngine) torrentsAsSlice() (torrents []*types.Torrent) {
	for _, torrent := range torrentEngine.torrents {
		torrents = append(torrents, torrent)
	}
	return
}
func (torrentEngine *TorrentEngine) GetTorrent(infohash string) (*types.Torrent, error) {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()
	torrent, ok := torrentEngine.torrents[infohash]
	if !ok {
		return nil, fmt.Errorf("torrent with infohash %s not found", infohash)
	}
	return torrent, nil
}
func (torrentEngine *TorrentEngine) RegisterTorrent(infohash string,
	name string,
	savePath string,
	specTrackers [][]string, addTopOfQueue bool) (*types.Torrent, error) {

	var err error

	// if save path empty use default path
	if savePath == "" {
		savePath = torrentEngine.Config.DownloadPath
	} else {
		if err = utils.CheckDirectoryExists(savePath); err != nil {
			return nil, err
		}
	}

	// Add trackers
	trackers := []types.Tracker{}
	for tierIndex, trackersOfTier := range specTrackers {
		for _, tracker := range trackersOfTier {
			//validate url
			trackerUrl, err := url.Parse(tracker)
			if err != nil {
				return nil, err
			}
			trackers = append(trackers, types.Tracker{
				Url:   trackerUrl.String(),
				Peers: []types.Peer{},
				Tier:  tierIndex,
			})
		}
	}

	dbTorrent := types.Torrent{
		Infohash: infohash,
		Name:     name,
		SavePath: savePath,
		Status:   types.TorrentStatusMetadata.String(),
		Trackers: trackers,
	}
	if addTopOfQueue {
		dbTorrent.QueueNumber = 1
	} else {
		if len(torrentEngine.torrents) == 0 {
			dbTorrent.QueueNumber = 1
		} else {
			lastQueueNumber, err := torrentEngine.db.GetLastQueueNumber()
			if err != nil {
				return nil, err
			}
			dbTorrent.QueueNumber = lastQueueNumber + 1
		}
	}
	// Insert torrent
	err = torrentEngine.db.InsertTorrent(&dbTorrent, addTopOfQueue)
	if err != nil {
		return nil, err
	}

	// Insert trackers
	for _, dbTracker := range dbTorrent.Trackers {
		if err = torrentEngine.db.InsertTracker(&dbTracker, dbTorrent.Infohash); err != nil {
			// if error is not 2067 (duplicate key) then return
			if sqliteErr, ok := err.(*sqlite.Error); ok && sqliteErr.Code() != 2067 {
				return nil, err
			}
		}
	}

	torrentEngine.mutexForTorrents.Lock()
	torrentEngine.torrents[dbTorrent.Infohash] = &dbTorrent
	torrentEngine.mutexForTorrents.Unlock()

	return &dbTorrent, nil
}

func (torrentEngine *TorrentEngine) AddTorrent(hash string, trackers []types.Tracker, savePath string, verifyFiles bool) (*gotorrent.Torrent, error) {
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
	torrent, new, err := torrentEngine.client.AddTorrentSpec(&torrentSpec)
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
func (torrentEngine *TorrentEngine) RegisterFiles(infohash metainfo.Hash, inputFiles *[]types.TorrentFileFlatTreeNode) (*types.Torrent, error) {
	torrent, ok := torrentEngine.client.Torrent(infohash)
	if !ok {
		return nil, fmt.Errorf("cannot find torrent with %s this infohash", infohash)
	}

	// torrentEngine.mutexForTorrents.Lock()
	// defer torrentEngine.mutexForTorrents.Unlock()
	dbTorrent := torrentEngine.torrents[infohash.String()]
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
				torrentEngine.db.InsertTorrentFile(&types.TorrentFileTreeNode{
					Path:     file.Path(),
					Priority: clientFile.Priority,
					Name:     fileName,
				}, dbTorrent.Infohash)
			}
		}

	}
	torrentEngine.db.UpdateSizeOfWanted(dbTorrent)
	dbTorrent.Files = torrentEngine.CreateFileTreeFromMeta(*torrent.Info())

	return dbTorrent, nil
}
func (torrentEngine *TorrentEngine) StartTorrent(torrent *gotorrent.Torrent) (*gotorrent.Torrent, error) {
	// set torrent file priorities
	// TODO(fatih): in the future we can make this a hashmap for faster search
	dbFiles, err := torrentEngine.db.GetTorrentTorrentFiles(torrent.InfoHash().String())
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
	torrentEngine.torrentPrevSizeMap[torrent.InfoHash().String()] = TorrentPrevSize{
		DownloadedBytes: torrent.BytesCompleted(),
		UploadedBytes:   0,
	}

	return torrent, nil
}
func (torrentEngine *TorrentEngine) FindTorrents(hashes []string) ([]*types.Torrent, error) {
	foundTorrents := []*types.Torrent{}
	for _, hash := range hashes {
		torrent, err := torrentEngine.GetTorrent(hash)
		if err != nil {
			return nil, fmt.Errorf("torrent with hash %s not found", hash)
		}
		foundTorrents = append(foundTorrents, torrent)
	}
	return foundTorrents, nil
}
func (torrentEngine *TorrentEngine) ResumeTorrent(hash string) error {
	clientTorrent, err := torrentEngine.getActiveClientTorrent(hash)
	if err != nil {
		return err
	}
	clientTorrent.SetMaxEstablishedConns(80)

	torrentEngine.UpdateTorrentStatus(hash, types.TorrentStatusDownloading)
	return nil
}
func (torrentEngine *TorrentEngine) PauseTorrent(hash string) error {
	clientTorrent, err := torrentEngine.getActiveClientTorrent(hash)
	if err != nil {
		return err
	}

	clientTorrent.CancelPieces(0, clientTorrent.NumPieces())
	clientTorrent.SetMaxEstablishedConns(0)

	torrentEngine.UpdateTorrentStatus(hash, types.TorrentStatusPaused)
	return nil
}
func (torrentEngine *TorrentEngine) RemoveTorrent(hash string) error {
	clientTorrent, err := torrentEngine.getActiveClientTorrent(hash)
	if err != nil {
		return err
	}

	torrentEngine.PauseTorrent(hash)
	clientTorrent.Drop()
	torrentEngine.db.DeleteTorrent(hash)

	err = torrentEngine.db.DeleteTorrentFilesByInfohash(hash)
	if err != nil {
		return err
	}
	err = torrentEngine.db.DeleteTorrentTrackerLinks(hash)
	if err != nil {
		return err
	}

	torrentEngine.mutexForTorrents.Lock()
	delete(torrentEngine.torrents, hash)
	torrentEngine.mutexForTorrents.Unlock()

	torrentEngine.updateTorrentQueueNumbers()
	return nil
}
func (torrentEngine *TorrentEngine) DeleteTorrent(hash string) error {
	clientTorrent, err := torrentEngine.getActiveClientTorrent(hash)
	if err != nil {
		return err
	}

	torrentEngine.mutexForTorrents.Lock()
	savePath := torrentEngine.torrents[hash].SavePath
	torrentEngine.mutexForTorrents.Unlock()

	err = torrentEngine.RemoveTorrent(hash)
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(savePath, clientTorrent.Name()))
	if err != nil {
		return err
	}
	return nil
}
func (torrentEngine *TorrentEngine) updateTorrentQueueNumbers() error {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()
	dbTorrents, err := torrentEngine.db.GetTorrents()
	if err != nil {
		return err
	}
	for _, dbTorrent := range dbTorrents {
		torrentEngine.torrents[dbTorrent.Infohash].QueueNumber = dbTorrent.QueueNumber
	}
	return nil
}
func (torrentEngine *TorrentEngine) getActiveClientTorrent(hash string) (*gotorrent.Torrent, error) {
	clientTorrent, ok := torrentEngine.client.Torrent(infohash.FromHexString(hash))
	if !ok {
		return nil, fmt.Errorf("cannot find torrent with %s this infohash", hash)
	}
	if clientTorrent.Info() == nil {
		return nil, fmt.Errorf("cannot modify torrent because metainfo is not yet received")
	}
	return clientTorrent, nil
}
func (torrentEngine *TorrentEngine) UpdateTorrentStatus(hash string, status types.TorrentStatus) error {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()

	torrentEngine.torrents[hash].Status = status.String()
	torrentEngine.db.UpdateTorrentStatus(hash, status)
	return nil
}
func (torrentEngine *TorrentEngine) GetTorrentDetails(infohash metainfo.Hash) (*types.Torrent, error) {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()

	dbTorrent, ok := torrentEngine.torrents[infohash.String()]
	if !ok {
		return nil, fmt.Errorf("torrent with infohash %s not found", infohash)
	}
	return dbTorrent, nil
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
func (torrentEngine *TorrentEngine) CreateFileTreeFromMeta(meta metainfo.Info) []*types.TorrentFileTreeNode {
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

func (torrentEngine *TorrentEngine) GetTotalDownloadSpeed() float32 {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()
	var totalDownloadSpeed float32
	for _, torrent := range torrentEngine.torrents {
		totalDownloadSpeed += torrent.DownloadSpeed
	}
	return totalDownloadSpeed
}
func (torrentEngine *TorrentEngine) GetTotalUploadSpeed() float32 {
	torrentEngine.mutexForTorrents.Lock()
	defer torrentEngine.mutexForTorrents.Unlock()
	var totalUploadSpeed float32
	for _, torrent := range torrentEngine.torrents {
		totalUploadSpeed += torrent.UploadSpeed
	}
	return totalUploadSpeed
}
func (torrentEngine *TorrentEngine) GetTorrentMetaWithMagnet(magnet string) (metainfo *metainfo.Info, infohash string, err error) {
	// Get metainfo from a magnet link
	torrent, err := torrentEngine.client.AddMagnet(magnet)
	if err != nil {
		return
	}

	<-torrent.GotInfo()

	metainfo = torrent.Info()
	infohash = torrent.InfoHash().String()

	torrent.Drop()
	return
}
