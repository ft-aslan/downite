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
type Response[T any] struct {
	Body T
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

type GetTorrentReq struct {
	Hash string `path:"hash" maxLength:"30" example:"2b66980093bc11806fab50cb3cb41835b95a0362" doc:"Hash of the torrent"`
}
type GetTorrentRes struct {
	Body types.Torrent
}

func GetTorrent(ctx context.Context, input *GetTorrentReq) (*GetTorrentRes, error) {
	res := &GetTorrentRes{}
	torrent, ok := torr.Client.Torrent(infohash.FromHexString(input.Hash))
	if !ok {
		return nil, fmt.Errorf("torrent with hash %s not found", input.Hash)
	}

	res.Body = types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}

	return res, nil
}

type DownloadTorrentReq struct {
	Body struct {
		Magnet                      string                     `json:"magnet"`
		TorrentFile                 []byte                     `json:"torrentFile"`
		SavePath                    string                     `json:"savePath" validate:"required, dir"`
		IsIncompleteSavePathEnabled bool                       `json:"isIncompleteSavePathEnabled"`
		IncompleteSavePath          string                     `json:"incompleteSavePath" validate:"dir"`
		Category                    string                     `json:"category,omitempty"`
		Tags                        []string                   `json:"tags,omitempty"`
		StartTorrent                bool                       `json:"startTorrent"`
		AddTopOfQueue               bool                       `json:"addTopOfQueue"`
		DownloadSequentially        bool                       `json:"downloadSequentially"`
		SkipHashCheck               bool                       `json:"skipHashCheck"`
		ContentLayout               string                     `json:"contentLayout" validate:"oneof='Original' 'Create subfolder' 'Don't create subfolder'"`
		Files                       []types.TorrentFileOptions `json:"files"`
	}
}
type DownloadTorrentRes struct {
	Body types.Torrent
}

func DownloadTorrent(ctx context.Context, input *DownloadTorrentReq) (*DownloadTorrentRes, error) {
	res := &DownloadTorrentRes{}
	var torrent *torrent.Torrent
	if input.Body.Magnet != "" {
		// Load from a magnet link
		torrent, err := torr.Client.AddMagnet(input.Body.Magnet)
		if err != nil {
			return nil, err
		}

		<-torrent.GotInfo()

	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(input.Body.TorrentFile)
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return nil, err
		}
		torrent, err = torr.Client.AddTorrent(torrentMeta)
		if err != nil {
			return nil, err
		}
	}
	if !input.Body.SkipHashCheck {
		torrent.VerifyData()
	}

	res.Body = types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}
	return res, nil
}

type GetTorrentMetaReq struct {
	Body struct {
		Magnet      string `json:"magnet,omitempty"`
		TorrentFile []byte `json:"torrentFile,omitempty"`
	}
}

type GetTorrentMetaRes struct {
	Body types.TorrentMeta
}

func GetTorrentMeta(ctx context.Context, input *GetTorrentMetaReq) (*GetTorrentMetaRes, error) {
	res := &GetTorrentMetaRes{}
	var info metainfo.Info
	if input.Body.Magnet != "" {
		// Load from a magnet link

		//torrent, err := metainfo.ParseMagnetUri(input.Body.Magnet)
		torrent, err := torr.Client.AddMagnet(input.Body.Magnet)
		if err != nil {
			return nil, err
		}

		<-torrent.GotInfo()

		info = *torrent.Info()
		torrent.Drop()
	} else {
		// Load the torrent file
		fileReader := bytes.NewReader(input.Body.TorrentFile)
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
	res.Body = types.TorrentMeta{
		TotalSize: info.TotalLength(),
		Files:     files,
		Name:      info.Name,
	}
	fmt.Printf("%s", ctx)
	return res, nil
}
