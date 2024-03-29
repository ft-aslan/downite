package handlers

import (
	"bytes"
	torr "downite/download/torrent"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/go-fuego/fuego"
)

type UploadTorrentReq struct {
	file   []byte
	magnet string
}
type UploadTorrentRes struct {
	totalSize int64
	files     metainfo.FileTree
	name      string
}

func GetTorrents(c *fuego.ContextNoBody) ([]torrent.Torrent, error) {
	return []torrent.Torrent{}, nil
}
func UploadTorrentMeta(c *fuego.ContextWithBody[UploadTorrentReq]) (UploadTorrentRes, error) {
	body, err := c.Body()
	if err != nil {
		return UploadTorrentRes{}, err
	}

	var info metainfo.Info
	if body.magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(body.magnet)
		if err != nil {
			return UploadTorrentRes{}, err
		}

		<-torrent.GotInfo()

		info = *torrent.Info()
		torrent.Drop()
	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(body.file)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return UploadTorrentRes{}, err
		}
		info, err = torrentMeta.UnmarshalInfo()
		if err != nil {
			return UploadTorrentRes{}, err
		}

	}

	return UploadTorrentRes{
		totalSize: info.TotalLength(),
		files:     info.FileTree,
		name:      info.Name,
	}, nil
}
