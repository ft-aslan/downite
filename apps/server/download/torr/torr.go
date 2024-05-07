package torr

import (
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

var Client *torrent.Client

func CreateTorrentClient(config *torrent.ClientConfig) {
	// Initialize the torrent client
	var err error
	Client, err = torrent.NewClient(config)
	if err != nil {
		fmt.Println("Error creating torrent client:", err)
		return
	}
}

func FindTorrents(hashes []string) ([]*torrent.Torrent, error) {
	foundTorrents := []*torrent.Torrent{}
	for _, hash := range hashes {
		torrent, ok := Client.Torrent(infohash.FromHexString(hash))
		if !ok {
			return nil, fmt.Errorf("torrent with hash %s not found", hash)
		}

		foundTorrents = append(foundTorrents, torrent)
	}
	return foundTorrents, nil
}
