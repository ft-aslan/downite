package handlers

import (
	"bytes"
	"downite/download/torr"
	"downite/types"
	"fmt"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/types/infohash"

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
func GetTorrent(c *fuego.ContextNoBody) (types.Torrent, error) {
	hash := c.PathParam("hash")
	if len(hash) < 20 {
		return types.Torrent{}, fmt.Errorf("invalid hash: %s", c.PathParam("hash"))
	}
	torrent, ok := torr.Client.Torrent(infohash.FromHexString(hash))
	if !ok {
		return types.Torrent{}, fmt.Errorf("torrent with hash %s not found", c.PathParam("hash"))
	}

	return types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}, nil
}

type DownloadTorrentReq struct {
	Magnet                      string           `json:"magnet" validate:"required_without=File, startswith=magnet:?"`
	File                        []byte           `json:"file" validate:"required_without=Magnet"`
	SavePath                    string           `json:"savePath" validate:"required, dir"`
	IsIncompleteSavePathEnabled bool             `json:"isIncompleteSavePathEnabled"`
	IncompleteSavePath          string           `json:"incompleteSavePath" validate:"dir"`
	Category                    string           `json:"category"`
	Tags                        []string         `json:"tags"`
	StartTorrent                bool             `json:"startTorrent"`
	AddTopOfQueue               bool             `json:"addTopOfQueue"`
	DownloadSequentially        bool             `json:"downloadSequentially"`
	SkipHashCheck               bool             `json:"skipHashCheck"`
	ContentLayout               string           `json:"contentLayout" validate:"oneof='Original' 'Create subfolder' 'Don't create subfolder'"`
	Files                       []types.FileMeta `json:"files"`
}

func DownloadTorrent(c *fuego.ContextWithBody[DownloadTorrentReq]) (types.Torrent, error) {
	var err error
	body, err := c.Body()
	if err != nil {
		return types.Torrent{}, err
	}
	var torrent *torrent.Torrent
	if body.Magnet != "" {
		// Load from a magnet link
		torrent, err = torr.Client.AddMagnet(body.Magnet)
		if err != nil {
			return types.Torrent{}, err
		}

		<-torrent.GotInfo()

	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(body.File)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return types.Torrent{}, err
		}
		torrent, err = torr.Client.AddTorrent(torrentMeta)
		if err != nil {
			return types.Torrent{}, err
		}
	}
	if !body.SkipHashCheck {
		torrent.VerifyData()
	}

	return types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}, nil
}

type GetTorrentMetaReq struct {
	File   []byte `json:"file"`
	Magnet string `json:"magnet"`
}
type TorrentMeta struct {
	TotalSize int64            `json:"totalSize"`
	Files     []types.FileMeta `json:"files"`
	Name      string           `json:"name"`
}

func GetTorrentMeta(c *fuego.ContextWithBody[GetTorrentMetaReq]) (TorrentMeta, error) {
	body, err := c.Body()
	if err != nil {
		return TorrentMeta{}, err
	}

	var info metainfo.Info
	if body.Magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(body.Magnet)
		if err != nil {
			return TorrentMeta{}, err
		}

		<-torrent.GotInfo()

		info = *torrent.Info()
		torrent.Drop()
	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(body.File)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return TorrentMeta{}, err
		}
		info, err = torrentMeta.UnmarshalInfo()
		if err != nil {
			return TorrentMeta{}, err
		}

	}
	var files []types.FileMeta
	for _, file := range info.Files {
		files = append(files, types.FileMeta{
			Length: file.Length,
			Path:   file.Path,
		})
	}
	return TorrentMeta{
		TotalSize: info.TotalLength(),
		Files:     files,
		Name:      info.Name,
	}, nil
}
