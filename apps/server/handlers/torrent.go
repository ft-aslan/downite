package handlers

import (
	"bytes"
	"context"
	"downite/download/torr"
	"downite/types"
	"fmt"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/types/infohash"
)

type GetTorrentsRes struct {
	Body struct {
		Torrents []types.Torrent `json:"torrents"`
	}
}

func GetTorrents(ctx context.Context, input *struct{}) (*GetTorrentsRes, error) {
	res := &GetTorrentsRes{}
	torrents := torr.Client.Torrents()

	var torrentsRes []types.Torrent
	for _, torrent := range torrents {

		torrentsRes = append(torrentsRes, types.Torrent{
			InfoHash: torrent.InfoHash().String(),
			Name:     torrent.Name(),
			AddedOn:  time.Now().Unix(),
			Files:    torrent.Info().FileTree,
		})
	}
	res.Body.Torrents = torrentsRes
	return res, nil
}
func GetTorrent(ctx context.Context, input *struct {
	Hash string `path:"hash" maxLength:"30" example:"2b66980093bc11806fab50cb3cb41835b95a0362" doc:"Hash of the torrent"`
}) (*types.Torrent, error) {

	torrent, ok := torr.Client.Torrent(infohash.FromHexString(input.Hash))
	if !ok {
		return nil, fmt.Errorf("torrent with hash %s not found", input.Hash)
	}

	return &types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}, nil
}

type DownloadTorrentReq struct {
	Magnet                      string           `json:"magnet"`
	File                        []byte           `json:"file"`
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

func DownloadTorrent(ctx context.Context, input *DownloadTorrentReq) (*types.Torrent, error) {
	var torrent *torrent.Torrent
	if input.Magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(input.Magnet)
		if err != nil {
			return nil, err
		}

		<-torrent.GotInfo()

	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(input.File)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return nil, err
		}
		torrent, err = torr.Client.AddTorrent(torrentMeta)
		if err != nil {
			return nil, err
		}
	}
	if !input.SkipHashCheck {
		torrent.VerifyData()
	}

	return &types.Torrent{
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

func GetTorrentMeta(ctx context.Context, input *GetTorrentMetaReq) (*TorrentMeta, error) {
	var info metainfo.Info
	if input.Magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(input.Magnet)
		if err != nil {
			return nil, err
		}

		<-torrent.GotInfo()

		info = *torrent.Info()
		torrent.Drop()
	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(input.File)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return nil, err
		}
		info, err = torrentMeta.UnmarshalInfo()
		if err != nil {
			return nil, err
		}

	}
	var files []types.FileMeta
	for _, file := range info.Files {
		files = append(files, types.FileMeta{
			Length: file.Length,
			Path:   file.Path,
		})
	}
	return &TorrentMeta{
		TotalSize: info.TotalLength(),
		Files:     files,
		Name:      info.Name,
	}, nil
}
