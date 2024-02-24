package main

import (
	"downite/download/direct_download"
	"fmt"
)

func main() {
	fmt.Println("Starting Downite server...")
	err := direct_download.DownloadFromUrl("https://i.redd.it/qh0xhmwhlakc1.jpeg", 8, "./")
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
