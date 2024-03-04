package handlers

import (
	"downite/download/torrent"

	"github.com/go-fuego/fuego"
)

func GetTorrents(c *fuego.ContextNoBody) ([]torrent.Torrent, error) {
	return []torrent.Torrent{}, nil
}
