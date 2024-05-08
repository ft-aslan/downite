package main

import (
	"downite/api"
	"downite/db"
	"downite/download/torr"
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

func main() {
	// Create a new torrent client
	torrentClientConfig := torrent.NewDefaultClientConfig()
	torrentClientConfig.DataDir = "./downloads"
	sqliteStorage, err := storage.NewSqlitePieceCompletion("./downloads")
	if err != nil {
		fmt.Printf("Error creating sqlite storage: %v\n", err)
		return
	}
	torrentClientConfig.DefaultStorage = storage.NewFileWithCompletion("./downloads", sqliteStorage)
	torr.CreateTorrentClient(torrentClientConfig)

	db.DbInit()
	api.ApiInit()
	// Load the torrent file
	// torrentInfo, err := metainfo.LoadFromFile("./mocks/debian-12.5.0-amd64-netinst.iso.torrent")
	// if err != nil {
	// 	fmt.Printf("Error reading torrent file: %v\n", err)
	// 	return
	// }
	// torrent, err := client.AddTorrent(torrentInfo)
	// if err != nil {
	// 	fmt.Printf("Error adding torrent: %v\n", err)
	// 	return
	// }
	// torrent.DownloadAll()
	// client.WaitAll()

	// err := direct_download.DownloadFromUrl("https://i.redd.it/qh0xhmwhlakc1.jpeg", 8, "./")
	// if err != nil {
	// 	fmt.Println("Error: ", err)
	// }
}
