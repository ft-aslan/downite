package main

import (
	"downite/api"
	"downite/db"
	"downite/download/protocol/torr"
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

}
