package main

import (
	"downite/api"
	"downite/db"
	"downite/download/protocol/direct"
	"downite/download/protocol/torr"
	"downite/handlers"
	"fmt"
	"os"
	"path/filepath"
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
	humaApi := api.ApiInit(api.ApiOptions{
		Port: 9999,
	})

	torrentEngine, err := torr.CreateTorrentEngine(torrentEngineConfig, db)
	if err != nil {
		fmt.Printf("Cannot create torrent engine : %s", err)
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

	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("Cannot get executable path : %s", err))
	}
	defaultDownloadsDir := filepath.Join(filepath.Dir(executablePath), "/tmp/downloads")
	// Check if the directory exists
	if _, err := os.Stat(defaultDownloadsDir); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(defaultDownloadsDir, os.ModePerm); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}
	downloadClientConfig := direct.DownloadClientConfig{
		DownloadPath: defaultDownloadsDir,
		PartCount:    8,
	}
	downloadClient, err := direct.CreateDownloadClient(downloadClientConfig, db)
	if err != nil {
		fmt.Printf("Cannot torrent download client : %s", err)
	}
	err = downloadClient.InitDownloads()
	if err != nil {
		fmt.Printf("Cannot initilize downloads : %s", err)
	}
	//register download routes
	humaApi.AddDownloadRoutes(handlers.DownloadHandler{
		Db:     db,
		Engine: downloadClient,
	})
	humaApi.ExportOpenApi()
	humaApi.Run()
}
