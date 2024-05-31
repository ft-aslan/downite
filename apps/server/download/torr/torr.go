package torr

import (
	"downite/db"
	"downite/types"
	"fmt"
	"path/filepath"
	"sort"
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
		go AddTorrent(&dbTorrent)
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

func AddTorrent(dbTorrent *types.Torrent) error {
	torrentSpec := gotorrent.TorrentSpec{
		InfoHash: infohash.FromHexString(dbTorrent.Infohash),
	}
	pieceCompletion, err := storage.NewDefaultPieceCompletionForDir("./tmp")
	if err != nil {
		return fmt.Errorf("new piece completion: %w", err)
	}
	torrentSpec.Storage = storage.NewFileOpts(storage.NewFileClientOpts{
		ClientBaseDir: dbTorrent.SavePath,
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
		return err
	}
	if !new {
		return fmt.Errorf("torrent with hash %s already exists", dbTorrent.Infohash)
	}
	// get the trackers
	trackers, err := db.GetTorrentTrackers(dbTorrent.Infohash)
	if err != nil {
		return err
	}
	if len(trackers) > 0 {
		// sort it based on their tiers
		sort.Slice(trackers, func(i, j int) bool { return trackers[i].Tier < trackers[j].Tier })
		// get the maximum tier number and create a tieredTrackers slice
		maximumTierIndex := trackers[len(trackers)-1].Tier
		tieredTrackers := make([][]string, 0, maximumTierIndex)
		// initialize the tieredTrackers slice
		for i := 0; i < maximumTierIndex+1; i++ {
			tieredTrackers = append(tieredTrackers, []string{})
		}
		// insert the trackers into the tieredTrackers slice based on their tiers
		for _, tracker := range trackers {
			tieredTrackers[tracker.Tier] = append(tieredTrackers[tracker.Tier], tracker.Url)
		}
		// Add trackers to the torrent
		torrent.AddTrackers(tieredTrackers)
	}

	// we need metainfo so we wait for it
	<-torrent.GotInfo()

	// verify the torrent
	torrent.VerifyData()

	// set torrent file priorities
	// TODO(fatih): in the future we can make this a hashmap for faster search
	dbFiles, err := db.GetTorrentTorrentFiles(torrent.InfoHash().String())
	if err != nil {
		return err
	}

	if dbTorrent.Status == types.TorrentStatusStringMap[types.TorrentStatusDownloading] {
		for _, file := range torrent.Files() {
			for _, dbFile := range dbFiles {
				if file.Path() == dbFile.Path {
					// set priority also starts the download for file if priority is not none
					file.SetPriority(types.PiecePriorityStringMap[dbFile.Priority])
				}
			}
		}
	}

	// get current size of torrent for speed calculation
	torrentPrevSizeMap[torrent.InfoHash().String()] = TorrentPrevSize{
		DownloadedBytes: torrent.BytesCompleted(),
		UploadedBytes:   0,
	}

	return nil
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
