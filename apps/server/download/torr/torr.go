package torr

import (
	"downite/db"
	"downite/types"
	"fmt"
	"sort"
	"sync"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

var Client *gotorrent.Client

var (
	lock             sync.Mutex
	torrentSpeedsMap = make(map[string]float64)
)

func CreateTorrentClient(config *gotorrent.ClientConfig) {
	// Initialize the torrent client
	var err error
	Client, err = gotorrent.NewClient(config)
	if err != nil {
		fmt.Println("Error creating torrent client:", err)
		return
	}
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

	return nil
}
func updateTorrentSpeeds() {
	torrents := Client.Torrents()
	for _, torrent := range torrents {
		stats := torrent.Stats()
		bytesWritten := stats.BytesWritten
		fmt.Printf("Torrent: %s - Speed: %f\n", torrent.InfoHash().HexString(), bytesWritten)
	}
}

func initTorrent(dbTorrent *types.Torrent) error {
	spec := gotorrent.TorrentSpec{
		InfoHash: infohash.FromHexString(dbTorrent.Infohash),
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

	if dbTorrent.Status == types.TorrentStatusDownloading {
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
