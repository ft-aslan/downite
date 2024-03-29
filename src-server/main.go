package main

import (
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func main() {
	// api.ApiInit()
	// db.DbInit()
	// Create a new torrent client
	torrentClientConfig := torrent.NewDefaultClientConfig()
	torrentClientConfig.DataDir = "./downloads"
	client, err := torrent.NewClient(torrentClientConfig)

	// Load the torrent file
	torrentInfo, err := metainfo.LoadFromFile("./mock/debian-12.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Printf("Error reading torrent file: %v\n", err)
		return
	}
	torrent, err := client.AddTorrent(torrentInfo)
	if err != nil {
		fmt.Printf("Error adding torrent: %v\n", err)
		return
	}
	torrent.DownloadAll()
	client.WaitAll()

	// err := direct_download.DownloadFromUrl("https://i.redd.it/qh0xhmwhlakc1.jpeg", 8, "./")
	// if err != nil {
	// 	fmt.Println("Error: ", err)
	// }
}
