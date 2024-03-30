package handlers

import (
	"bytes"
	"downite/download/torr"
	"downite/types"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"github.com/go-fuego/fuego"
)

func GetTorrents(c *fuego.ContextNoBody) ([]types.Torrent, error) {
	torrents := torr.Client.Torrents()

	var res []types.Torrent
	for _, torrent := range torrents {

		res = append(res, types.Torrent{
			InfoHash: torrent.InfoHash().String(),
			Name:     torrent.Name(),
			AddedOn:  time.Now().Unix(),
			Files:    torrent.Info().FileTree,
		})
	}
	return res, nil
}

type GetTorrentMetaReq struct {
	File   []byte `json:"file"`
	Magnet string `json:"magnet"`
}
type GetTorrentMetaRes struct {
	TotalSize int64            `json:"totalSize"`
	Files     []types.FileMeta `json:"files"`
	Name      string           `json:"name"`
}

func GetTorrentMeta(c *fuego.ContextWithBody[GetTorrentMetaReq]) (GetTorrentMetaRes, error) {
	body, err := c.Body()
	if err != nil {
		return GetTorrentMetaRes{}, err
	}

	var info metainfo.Info
	if body.Magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(body.Magnet)
		if err != nil {
			return GetTorrentMetaRes{}, err
		}

		<-torrent.GotInfo()

		info = *torrent.Info()
		torrent.Drop()
	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(body.File)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return GetTorrentMetaRes{}, err
		}
		info, err = torrentMeta.UnmarshalInfo()
		if err != nil {
			return GetTorrentMetaRes{}, err
		}

	}
	var files []types.FileMeta
	for _, file := range info.Files {
		files = append(files, types.FileMeta{
			Length: file.Length,
			Path:   file.Path,
		})
	}
	return GetTorrentMetaRes{
		TotalSize: info.TotalLength(),
		Files:     files,
		Name:      info.Name,
	}, nil
}
