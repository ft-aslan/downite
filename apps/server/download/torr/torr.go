package torr

import (
	"downite/db"
	"fmt"

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
		_, new, err := Client.AddTorrentSpec(&spec)

		if err != nil {
			return err
		}
		if !new {
			return fmt.Errorf("torrent with hash %s already exists", dbTorrent.Infohash)
		}
		// Add trackers to the torrent
		// _, err = torrent.AddTrackers(dbTorrent.Trackers)

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
