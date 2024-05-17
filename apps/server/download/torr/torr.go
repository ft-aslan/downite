package torr

import (
	"downite/db"
	"downite/types"
	"fmt"
	"sort"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

var Client *gotorrent.Client

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
			// insert the trackers into the tieredTrackers slice based on their tiers
			for _, tracker := range trackers {
				tieredTrackers[tracker.Tier] = append(tieredTrackers[tracker.Tier], tracker.Url.String())
			}
			// Add trackers to the torrent
			torrent.AddTrackers(tieredTrackers)
		}
		if dbTorrent.Status == types.TorrentStatusDownloading {
			torrent.DownloadAll()
		}
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
