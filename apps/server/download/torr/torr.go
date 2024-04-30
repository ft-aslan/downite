package torr

import (
	"fmt"

	"github.com/anacrolix/torrent"
)

var Client *torrent.Client

func CreateTorrentClient(config *torrent.ClientConfig) {
	// Initialize the torrent client
	var err error
	Client, err = torrent.NewClient(torrent.NewDefaultClientConfig())
	if err != nil {
		fmt.Println("Error creating torrent client:", err)
		return
	}
}
