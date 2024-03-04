package main

import (
	"downite/download/torrent"
	"fmt"
)

func main() {
	// api.ApiInit()
	// db.DbInit()
	t, err := torrent.New("./mock/debian-12.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	t.DownloadTorrent()
	// err := direct_download.DownloadFromUrl("https://i.redd.it/qh0xhmwhlakc1.jpeg", 8, "./")
	// if err != nil {
	// 	fmt.Println("Error: ", err)
	// }
}
