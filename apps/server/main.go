package main

import (
	"downite/api"
	"downite/db"
	"downite/download/protocol/http"
	"downite/download/protocol/torr"
	"downite/handlers"
	"fmt"
)

func main() {
	pieceCompletionDir := "./tmp"
	defaultTorrentsDir := "./tmp/torrents"
	torrentEngineConfig := torr.TorrentEngineConfig{
		PieceCompletionDbPath: pieceCompletionDir,
		DownloadPath:          defaultTorrentsDir,
	}
	db, err := db.DbInit()
	if err != nil {
		fmt.Printf("Cannot connect to db : %s", err)
	}
	humaApi := api.ApiInit(api.ApiOptions{})

	torrentEngine, err := torr.CreateTorrentEngine(torrentEngineConfig, db)
	if err != nil {
		fmt.Printf("Cannot torrent download client : %s", err)
	}
	err = torrentEngine.InitTorrents()
	if err != nil {
		fmt.Printf("Cannot initilize torrents : %s", err)
	}
	//register torrent routes
	humaApi.AddTorrentRoutes(handlers.TorrentHandler{
		Db:     db,
		Engine: torrentEngine,
	})

	defaultDownloadsDir := "./tmp/downloads"
	downloadClientConfig := http.DownloadClientConfig{
		DownloadPath: defaultDownloadsDir,
	}
	downloadClient, err := http.CreateDownloadClient(downloadClientConfig)
	if err != nil {
		fmt.Printf("Cannot torrent download client : %s", err)
	}
	err = torrentEngine.InitTorrents()
	if err != nil {
		fmt.Printf("Cannot initilize torrents : %s", err)
	}
	//register download routes
	humaApi.AddDownloadRoutes(handlers.DownloadHandler{
		Db:     db,
		Client: downloadClient,
	})
}
