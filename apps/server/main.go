package main

import (
	"downite/api"
	"downite/db"
	"downite/download/torr"
	"downite/types"
)

func main() {
	pieceCompletionDir := "./tmp"
	defaultTorrentsDir := "./tmp/downloads"
	torrentClientConfig := types.TorrentClientConfig{
		PieceCompletionDbPath: pieceCompletionDir,
		DownloadPath:          defaultTorrentsDir,
	}
	db.DbInit()
	torr.CreateTorrentClient(torrentClientConfig)
	torr.InitTorrents()
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
