package torr

import (
	"downite/db"
	"downite/types"
	"fmt"
	"sort"
	"sync"
	"time"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
)

var Client *gotorrent.Client

var (
	MutexForTorrentSpeed sync.Mutex
	TorrentSpeedMap      = make(map[string]float32)
	torrentPrevSizeMap   = make(map[string]int64)
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
		go initTorrent(&dbTorrent)
	}

	// Start a goroutine to update download speed
	go updateTorrentSpeeds()
	return nil
}
func updateTorrentSpeeds() {
	for {
		torrents := Client.Torrents()
		for _, torrent := range torrents {
			// stats := torrent.Stats()
			MutexForTorrentSpeed.Lock()
			prevTotalLength := torrentPrevSizeMap[torrent.InfoHash().HexString()]
			newTotalLength := torrent.BytesCompleted()
			downloadedByteCount := newTotalLength - prevTotalLength
			downloadSpeed := float32(downloadedByteCount) / 1024
			TorrentSpeedMap[torrent.InfoHash().HexString()] = downloadSpeed
			torrentPrevSizeMap[torrent.InfoHash().HexString()] = newTotalLength
			MutexForTorrentSpeed.Unlock()
		}
		time.Sleep(time.Second)
	}

}

func initTorrent(dbTorrent *types.Torrent) error {
	spec := gotorrent.TorrentSpec{
		InfoHash: infohash.FromHexString(dbTorrent.Infohash),
		Storage:  storage.NewFile(dbTorrent.SavePath),
	}
	torrent, new, err := Client.AddTorrentSpec(&spec)

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
	for _, file := range torrent.Files() {
		for _, dbFile := range dbFiles {
			if file.Path() == dbFile.Path {
				file.SetPriority(types.PiecePriorityStringMap[dbFile.Priority])
			}
		}
	}

	// get current size of torrent for speed calculation
	torrentPrevSizeMap[torrent.InfoHash().String()] = torrent.BytesCompleted()

	if dbTorrent.Status == types.TorrentStatusStringMap[types.TorrentStatusDownloading] {
		torrent.DownloadAll()
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
