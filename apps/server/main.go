package main

import (
	"downite/api"
	"downite/db"
	"downite/download/protocol/torr"
	"downite/types"
	"fmt"
)

func main() {
	pieceCompletionDir := "./tmp"
	defaultTorrentsDir := "./tmp/downloads"
	torrentClientConfig := types.TorrentClientConfig{
		PieceCompletionDbPath: pieceCompletionDir,
		DownloadPath:          defaultTorrentsDir,
	}
	db, err := db.DbInit()
	if err != nil {
		fmt.Printf("Cannot connect to db : %s", err)
	}
	torrentClient, err := torr.CreateTorrentClient(torrentClientConfig)
	torr.InitTorrents()
	api.ApiInit()

}
