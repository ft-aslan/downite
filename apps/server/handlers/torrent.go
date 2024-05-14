package handlers

import (
	"bytes"
	"context"
	"downite/db"
	"downite/download/torr"
	"downite/types"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/jmoiron/sqlx"
)

type GetTorrentsRes struct {
	Body struct {
		Torrents []types.Torrent `json:"torrents"`
	}
}

func GetTorrents(ctx context.Context, input *struct{}) (*GetTorrentsRes, error) {
	res := &GetTorrentsRes{}
	torrentsRes := []types.Torrent{}

	/* 	torrents := torr.Client.Torrents()
	   	for _, torrent := range torrents {
	   		if torrent.Info() == nil {
	   			torrentsRes = append(torrentsRes, types.Torrent{
	   				InfoHash: torrent.InfoHash().String(),
	   				Name:     torrent.Name(),
	   				AddedOn:  time.Now().Unix(),
	   				Status:   "loading",
	   			})
	   		} else {
	   			torrentsRes = append(torrentsRes, types.Torrent{
	   				InfoHash:   torrent.InfoHash().String(),
	   				Name:       torrent.Name(),
	   				AddedOn:    time.Now().Unix(),
	   				Files:      torrent.Info().FileTree,
	   				TotalSize:  torrent.Info().TotalLength(),
	   				AmountLeft: torrent.BytesMissing(),
	   				Downloaded: torrent.BytesCompleted(),
	   				Progress:   float32(torrent.BytesCompleted()) / float32(torrent.Info().TotalLength()) * 100,
	   				Status:     "downloading",
	   			})
	   		}
	   	} */
	torrents, err := db.GetTorrents()
	if err != nil {
		return nil, err
	}
	for _, torrent := range torrents {
		torrentsRes = append(torrentsRes, types.Torrent{
			InfoHash:   torrent.InfoHash,
			Name:       torrent.Name,
			AddedOn:    torrent.AddedOn,
			Files:      torrent.Files,
			TotalSize:  torrent.TotalSize,
			AmountLeft: torrent.AmountLeft,
			Downloaded: torrent.Downloaded,
			Progress:   torrent.Progress,
			Status:     torrent.Status,
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

type TorrentActionReq struct {
	Body struct {
		InfoHashes []string `json:"infoHashes" maxLength:"30" example:"2b66980093bc11806fab50cb3cb41835b95a0362" doc:"Hashes of torrents"`
	}
}
type TorrentActionRes struct {
	Body struct {
		Success bool `json:"result"`
	}
}

func PauseTorrent(ctx context.Context, input *TorrentActionReq) (*TorrentActionRes, error) {
	res := &TorrentActionRes{}
	foundTorrents, err := torr.FindTorrents(input.Body.InfoHashes)
	if err != nil {
		return nil, err
	}
	for _, torrent := range foundTorrents {
		if torrent.Info() != nil {
			torrent.CancelPieces(0, torrent.NumPieces())
			sqlx.MustExec(db.DB, "UPDATE torrents SET status = ? WHERE infohash = ?", types.TorrentStatusPaused, torrent.InfoHash().String())
		} else {
			return nil, fmt.Errorf("cannot modify torrent because metainfo is not yet received")
		}
	}
	res.Body.Success = true

	return res, nil
}
func ResumeTorrent(ctx context.Context, input *TorrentActionReq) (*TorrentActionRes, error) {
	res := &TorrentActionRes{}
	foundTorrents, err := torr.FindTorrents(input.Body.InfoHashes)
	if err != nil {
		return nil, err
	}
	for _, foundTorrent := range foundTorrents {
		// TODO(fatih): check if torrent is already started
		if foundTorrent.Info() != nil {
			foundTorrent.DownloadAll()
			sqlx.MustExec(db.DB, "UPDATE torrents SET status = ? WHERE infohash = ?", types.TorrentStatusDownloading, foundTorrent.InfoHash().String())
		} else {
			return nil, fmt.Errorf("cannot modify torrent because metainfo is not yet received")
		}
	}
	res.Body.Success = true

	return res, nil
}
func RemoveTorrent(ctx context.Context, input *TorrentActionReq) (*TorrentActionRes, error) {
	res := &TorrentActionRes{}
	foundTorrents, err := torr.FindTorrents(input.Body.InfoHashes)
	if err != nil {
		return nil, err
	}
	for _, foundTorrent := range foundTorrents {
		// TODO(fatih): check if torrent is already started
		if foundTorrent.Info() != nil {
			foundTorrent.CancelPieces(0, foundTorrent.NumPieces())
			foundTorrent.Drop()
			sqlx.MustExec(db.DB, "DELETE FROM torrents WHERE infohash = ?", foundTorrent.InfoHash().String())
		} else {
			return nil, fmt.Errorf("cannot modify torrent because metainfo is not yet received")
		}
	}
	res.Body.Success = true

	return res, nil
}

type DownloadTorrentReq struct {
	Body struct {
		Magnet                      string                     `json:"magnet,omitempty"`
		TorrentFile                 string                     `json:"torrentFile,omitempty"`
		SavePath                    string                     `json:"savePath" validate:"required, dir"`
		IsIncompleteSavePathEnabled bool                       `json:"isIncompleteSavePathEnabled"`
		IncompleteSavePath          string                     `json:"incompleteSavePath,omitempty" validate:"dir"`
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
	var err error
	if input.Body.Magnet != "" {
		// Load from a magnet link
		torrent, err = torr.Client.AddMagnet(input.Body.Magnet)
		if err != nil {
			return nil, err
		}

	} else {
		// Load the torrent file
		fileReader := bytes.NewReader([]byte(input.Body.TorrentFile))
		torrentMeta, err := metainfo.Load(fileReader)
		if err != nil {
			return nil, err
		}
		torrent, err = torr.Client.AddTorrent(torrentMeta)
		if err != nil {
			return nil, err
		}
	}
	sqlx.MustExec(db.DB,
		"INSERT INTO torrents (infohash, name, created_at, save_path, status, time_active, downloaded, uploaded, total_size, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		torrent.InfoHash().String(), torrent.Name(), time.Now(), input.Body.SavePath, types.TorrentStatusMetadata, 0, 0, 0, 0, "")

	<-torrent.GotInfo()

	sqlx.MustExec(db.DB, "UPDATE torrents SET status = ? WHERE infohash = ?",
		types.TorrentStatusPaused, torrent.InfoHash().String())

	for _, file := range torrent.Files() {
		for _, clientFile := range input.Body.Files {
			if file.Path() == clientFile.Path {
				priority, ok := types.PiecePriorityStringMap[clientFile.DownloadPriority]
				if !ok {
					return nil, fmt.Errorf("invalid download priority: %s", clientFile.DownloadPriority)
				}
				file.SetPriority(priority)
			}
		}

	}

	if !input.Body.SkipHashCheck {
		torrent.VerifyData()
	}

	if input.Body.StartTorrent {
		torrent.DownloadAll()
		sqlx.MustExec(db.DB, "UPDATE torrents SET status = ? WHERE infohash = ?",
			types.TorrentStatusDownloading, torrent.InfoHash().String())
	}

	res.Body = types.Torrent{
		InfoHash: torrent.InfoHash().String(),
		Name:     torrent.Name(),
		AddedOn:  time.Now().Unix(),
		Files:    torrent.Info().FileTree,
	}
	return res, nil
}

type GetMetaWithFileReq struct {
	RawBody multipart.Form
}

type GetMetaWithFileRes struct {
	Body types.TorrentMeta
}

func GetMetaWithFile(ctx context.Context, input *GetMetaWithFileReq) (*GetMetaWithFileRes, error) {
	// TODO(fatih): In the future, we should support multiple torrents
	res := &GetMetaWithFileRes{}
	torrentFiles := input.RawBody.File["torrentFile"]

	// Form validation
	if len(torrentFiles) == 0 {
		return nil, errors.New("no torrent file provided")
	}
	if len(torrentFiles) > 1 {
		return nil, errors.New("only one torrent file can be provided")
	}

	var info metainfo.Info
	var infoHash string
	var magnet string

	// Load the torrent file
	torrentFile, err := torrentFiles[0].Open()
	if err != nil {
		return nil, err
	}
	defer torrentFile.Close()

	torrentMeta, err := metainfo.Load(torrentFile)
	if err != nil {
		return nil, err
	}
	info, err = torrentMeta.UnmarshalInfo()
	infoHash = torrentMeta.HashInfoBytes().String()
	magnetInfo := torrentMeta.Magnet(nil, &info)
	magnet = magnetInfo.String()
	if err != nil {
		return nil, err
	}

	fileTree := createFileTreeFromMeta(info)

	res.Body = types.TorrentMeta{
		TotalSize:     info.TotalLength(),
		Files:         fileTree,
		Name:          info.Name,
		InfoHash:      infoHash,
		TorrentMagnet: magnet,
	}
	return res, nil
}

type GetMetaWithMagnetReq struct {
	Body struct {
		Magnet string `json:"magnet" minLength:"1"`
	}
}

type GetMetaWithMagnetRes struct {
	Body types.TorrentMeta
}

func GetMetaWithMagnet(ctx context.Context, input *GetMetaWithMagnetReq) (*GetMetaWithMagnetRes, error) {
	// TODO(fatih): In the future, we should support multiple torrents
	res := &GetMetaWithMagnetRes{}

	var info metainfo.Info
	var infoHash string

	magnet := input.Body.Magnet
	if _, err := metainfo.ParseMagnetUri(magnet); err != nil {
		return nil, errors.New("invalid magnet")
	}
	// Load from a magnet link

	torrent, err := torr.Client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	<-torrent.GotInfo()

	info = *torrent.Info()
	infoHash = torrent.InfoHash().String()

	fileTree := createFileTreeFromMeta(info)

	res.Body = types.TorrentMeta{
		TotalSize:     info.TotalLength(),
		Files:         fileTree,
		Name:          info.Name,
		InfoHash:      infoHash,
		TorrentMagnet: magnet,
	}

	torrent.Drop()
	return res, nil
}
func createFolder(fileTree *[]*types.TreeNodeMeta, path []string) (*[]*types.TreeNodeMeta, *types.TreeNodeMeta) {
	currentFileTree := fileTree
	var parentNode *types.TreeNodeMeta
	for pathIndex, segment := range path {
		currentPath := path[:pathIndex+1]
		found := false
		if len(*currentFileTree) > 0 {
			for _, node := range *currentFileTree {
				if node.Name == segment {
					parentNode = node
					currentFileTree = node.Children
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		parentNode = &types.TreeNodeMeta{
			Length:   0,
			Name:     segment,
			Path:     currentPath,
			Children: &[]*types.TreeNodeMeta{},
		}
		*currentFileTree = append(*currentFileTree, parentNode)
		currentFileTree = parentNode.Children
	}

	return currentFileTree, parentNode
}
func createFileTreeFromMeta(meta metainfo.Info) []*types.TreeNodeMeta {
	var fileTree []*types.TreeNodeMeta
	//there is no file tree in torrent
	if len(meta.Files) == 0 {
		fileTree = []*types.TreeNodeMeta{
			{
				Length:   meta.TotalLength(),
				Name:     meta.Name,
				Path:     []string{meta.Name},
				Children: &[]*types.TreeNodeMeta{},
			},
		}
	}
	//there is a file tree in torrent
	for _, file := range meta.Files {
		targetNodeTree := &fileTree
		var parentNode *types.TreeNodeMeta
		if len(file.Path) > 1 {
			targetNodeTree, parentNode = createFolder(targetNodeTree, file.Path[:len(file.Path)-1])
		}
		*targetNodeTree = append(*targetNodeTree, &types.TreeNodeMeta{
			Length:   file.Length,
			Name:     file.Path[len(file.Path)-1],
			Path:     file.Path,
			Children: &[]*types.TreeNodeMeta{},
		})
		if parentNode != nil {
			parentNode.Length += file.Length
		}
	}
	return fileTree
}
