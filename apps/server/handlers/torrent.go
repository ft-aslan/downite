package handlers

import (
	"bytes"
	"context"
	"downite/db"
	"downite/download/torr"
	"downite/types"
	"downite/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	gotorrenttypes "github.com/anacrolix/torrent/types"
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

	torrents := torr.Client.Torrents()
	for _, torrent := range torrents {
		dbTorrent, err := db.GetTorrent(torrent.InfoHash().String())
		if err != nil {
			return nil, err
		}
		if torrent.Info() == nil {
			torrentsRes = append(torrentsRes, types.Torrent{
				Infohash:  torrent.InfoHash().String(),
				Name:      torrent.Name(),
				CreatedAt: dbTorrent.CreatedAt,
				Status:    types.TorrentStatusStringMap[types.TorrentStatusMetadata],
			})
		} else {

			fileTree := createFileTreeFromMeta(*torrent.Info())
			newTorrent := types.Torrent{
				Infohash:   torrent.InfoHash().String(),
				Name:       torrent.Name(),
				CreatedAt:  dbTorrent.CreatedAt,
				Files:      fileTree,
				TotalSize:  torrent.Info().TotalLength(),
				AmountLeft: torrent.BytesMissing(),
				Downloaded: torrent.BytesCompleted(),
				Progress:   float32(torrent.BytesCompleted()) / float32(dbTorrent.SizeOfWanted) * 100,
				Seeds:      torrent.Stats().ConnectedSeeders,
				PeerCount:  torrent.Stats().ActivePeers,
				Status:     dbTorrent.Status,
			}

			// we use mutex becouse calculating speed is concurrent
			torr.MutexForTorrentSpeed.Lock()
			speeds := torr.TorrentSpeedMap[torrent.InfoHash().String()]
			newTorrent.DownloadSpeed = speeds.DownloadSpeed
			newTorrent.UploadSpeed = speeds.UploadSpeed
			torr.MutexForTorrentSpeed.Unlock()

			torrentsRes = append(torrentsRes, newTorrent)
		}
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
		Infohash:  torrent.InfoHash().String(),
		Name:      torrent.Name(),
		CreatedAt: time.Now().Unix(),
		// Files:     torrent.Info().FileTree,
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
	for _, foundTorrent := range foundTorrents {
		if foundTorrent.Info() != nil {
			foundTorrent.CancelPieces(0, foundTorrent.NumPieces())
			torrent, err := db.GetTorrent(foundTorrent.InfoHash().String())
			if err != nil {
				return nil, err
			}
			torrent.Status = types.TorrentStatusStringMap[types.TorrentStatusPaused]
			db.UpdateTorrentStatus(torrent)

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

			torrent, err := db.GetTorrent(foundTorrent.InfoHash().String())
			if err != nil {
				return nil, err
			}
			torrent.Status = types.TorrentStatusStringMap[types.TorrentStatusDownloading]
			db.UpdateTorrentStatus(torrent)

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

// this is also deletes the torrent from disk
func DeleteTorrent(ctx context.Context, input *TorrentActionReq) (*TorrentActionRes, error) {
	res := &TorrentActionRes{}
	foundTorrents, err := torr.FindTorrents(input.Body.InfoHashes)
	if err != nil {
		return nil, err
	}
	for _, foundTorrent := range foundTorrents {
		// TODO(fatih): check if torrent is already started
		if foundTorrent.Info() == nil {
			return nil, fmt.Errorf("cannot modify torrent because metainfo is not yet received")
		}
		foundTorrent.CancelPieces(0, foundTorrent.NumPieces())
		foundTorrent.Drop()

		dbTorrent, err := db.GetTorrent(foundTorrent.InfoHash().String())
		err = db.DeleteTorrent(foundTorrent.InfoHash().String())
		err = db.DeleteTorrentFilesByInfohash(foundTorrent.InfoHash().String())
		err = os.RemoveAll(filepath.Join(dbTorrent.SavePath, foundTorrent.Name()))
		if err != nil {
			return nil, err
		}
	}
	res.Body.Success = true
	return res, nil
}

type DownloadTorrentReq struct {
	Body struct {
		Magnet                      string                          `json:"magnet,omitempty"`
		TorrentFile                 string                          `json:"torrentFile,omitempty"`
		SavePath                    string                          `json:"savePath" validate:"required, dir"`
		IsIncompleteSavePathEnabled bool                            `json:"isIncompleteSavePathEnabled"`
		IncompleteSavePath          string                          `json:"incompleteSavePath,omitempty" validate:"dir"`
		Category                    string                          `json:"category,omitempty"`
		Tags                        []string                        `json:"tags,omitempty"`
		StartTorrent                bool                            `json:"startTorrent"`
		AddTopOfQueue               bool                            `json:"addTopOfQueue"`
		DownloadSequentially        bool                            `json:"downloadSequentially"`
		SkipHashCheck               bool                            `json:"skipHashCheck"`
		ContentLayout               string                          `json:"contentLayout" enum:"Original,Create subfolder,Don't create subfolder"`
		Files                       []types.TorrentFileFlatTreeNode `json:"files"`
	}
}
type DownloadTorrentRes struct {
	Body types.Torrent
}

func DownloadTorrent(ctx context.Context, input *DownloadTorrentReq) (*DownloadTorrentRes, error) {
	res := &DownloadTorrentRes{}
	var torrent *gotorrent.Torrent
	var torrentSpec *gotorrent.TorrentSpec
	var dbTorrent types.Torrent
	var dbTrackers []types.Tracker
	var savePath string

	var err error
	if input.Body.Magnet != "" {
		// Validate magnet
		if _, err = metainfo.ParseMagnetUri(input.Body.Magnet); err != nil {
			return nil, errors.New("invalid magnet")
		}

		// Load from a magnet link
		torrentSpec, err = gotorrent.TorrentSpecFromMagnetUri(input.Body.Magnet)
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
		torrentSpec = gotorrent.TorrentSpecFromMetaInfo(torrentMeta)

	}
	if torrentSpec == nil {
		return nil, errors.New("invalid torrent")
	}

	specTrackers := torrentSpec.Trackers
	for tierIndex, trackersOfTier := range specTrackers {
		for _, tracker := range trackersOfTier {
			//validate url
			trackerUrl, err := url.Parse(tracker)
			if err != nil {
				return nil, err
			}
			dbTrackers = append(dbTorrent.Trackers, types.Tracker{
				Url:  trackerUrl.String(),
				Tier: tierIndex,
			})
		}
	}

	savePath = input.Body.SavePath
	// if save path empty use default path
	if savePath == "" {
		savePath = torr.TorrentClientConfig.DownloadPath
	} else {
		if err = utils.CheckDirectoryExists(savePath); err != nil {
			return nil, err
		}
	}

	dbTorrent = types.Torrent{
		Infohash: torrentSpec.InfoHash.String(),
		Name:     torrentSpec.DisplayName,
		SavePath: savePath,
		Status:   types.TorrentStatusStringMap[types.TorrentStatusMetadata],
		Trackers: dbTrackers,
	}

	err = db.InsertTorrent(&dbTorrent)
	if err != nil {
		return nil, err
	}
	for _, dbTracker := range dbTorrent.Trackers {
		if err = db.InsertTracker(&dbTracker, dbTorrent.Infohash); err != nil {
			return nil, err
		}
	}
	// ADD TORRENT TO CLIENT
	torrent, err = torr.AddTorrent(&dbTorrent, input.Body.StartTorrent, !input.Body.SkipHashCheck)
	if err != nil {
		return nil, err
	}

	dbTorrent.Status = types.TorrentStatusStringMap[types.TorrentStatusPaused]
	dbTorrent.TotalSize = torrent.Length()
	dbTorrent.Magnet = torrent.Metainfo().Magnet(nil, torrent.Info()).String()

	db.UpdateTorrent(&dbTorrent)

	// Insert trackers
	trackers := torrent.Metainfo().AnnounceList
	for tierIndex, trackersOfTier := range trackers {
		for _, tracker := range trackersOfTier {
			//validate url
			trackerUrl, err := url.Parse(tracker)
			if err != nil {
				return nil, err
			}
			db.InsertTracker(&types.Tracker{
				Url:  trackerUrl.String(),
				Tier: tierIndex,
			}, dbTorrent.Infohash)
		}
	}

	// Set download priorities of the files
	for _, file := range torrent.Files() {
		for _, clientFile := range input.Body.Files {
			if file.DisplayPath() == clientFile.Path {
				priority, ok := types.PiecePriorityStringMap[clientFile.Priority]
				if !ok {
					return nil, fmt.Errorf("invalid download priority: %s", clientFile.Priority)
				}

				if input.Body.StartTorrent {
					// set priority also starts the download for file if priority is not none
					file.SetPriority(priority)
				}
				if priority != gotorrenttypes.PiecePriorityNone {
					dbTorrent.SizeOfWanted += file.Length()
				}

				db.InsertTorrentFile(&types.TorrentFileTreeNode{
					Path:     file.Path(),
					Priority: clientFile.Priority,
					Name:     file.FileInfo().Path[len(file.FileInfo().Path)-1],
				}, dbTorrent.Infohash)
			}
		}

	}

	// update size of wanted in torrent
	db.UpdateTorrent(&dbTorrent)

	if !input.Body.SkipHashCheck {
		torrent.VerifyData()
	}

	if input.Body.StartTorrent {
		dbTorrent.Status = types.TorrentStatusStringMap[types.TorrentStatusDownloading]
		db.UpdateTorrentStatus(&dbTorrent)
	}

	res.Body = dbTorrent

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
	var infohash string
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
	infohash = torrentMeta.HashInfoBytes().String()
	magnetInfo := torrentMeta.Magnet(nil, &info)
	magnet = magnetInfo.String()
	if err != nil {
		return nil, err
	}

	fileTree := createFileTreeFromMeta(info)

	res.Body = types.TorrentMeta{
		TotalSize: info.TotalLength(),
		Files:     fileTree,
		Name:      info.Name,
		Infohash:  infohash,
		Magnet:    magnet,
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
	var infohash string

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
	infohash = torrent.InfoHash().String()

	fileTree := createFileTreeFromMeta(info)

	res.Body = types.TorrentMeta{
		TotalSize: info.TotalLength(),
		Files:     fileTree,
		Name:      info.Name,
		Infohash:  infohash,
		Magnet:    magnet,
	}

	torrent.Drop()
	return res, nil
}
func createFolder(fileTree *[]*types.TorrentFileTreeNode, path []string) (*[]*types.TorrentFileTreeNode, *types.TorrentFileTreeNode) {
	currentFileTree := fileTree
	var parentNode *types.TorrentFileTreeNode
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
		parentNode = &types.TorrentFileTreeNode{
			Length:   0,
			Name:     segment,
			Path:     strings.Join(currentPath, "/"),
			Children: &[]*types.TorrentFileTreeNode{},
		}
		*currentFileTree = append(*currentFileTree, parentNode)
		currentFileTree = parentNode.Children
	}

	return currentFileTree, parentNode
}
func createFileTreeFromMeta(meta metainfo.Info) []*types.TorrentFileTreeNode {
	var fileTree []*types.TorrentFileTreeNode
	//there is no file tree in torrent
	if len(meta.Files) == 0 {
		fileTree = []*types.TorrentFileTreeNode{
			{
				Length:   meta.TotalLength(),
				Name:     meta.Name,
				Path:     meta.Name,
				Children: &[]*types.TorrentFileTreeNode{},
			},
		}
	}
	//there is a file tree in torrent
	for _, file := range meta.Files {
		targetNodeTree := &fileTree
		var parentNode *types.TorrentFileTreeNode
		if len(file.Path) > 1 {
			targetNodeTree, parentNode = createFolder(targetNodeTree, file.Path[:len(file.Path)-1])
		}
		*targetNodeTree = append(*targetNodeTree, &types.TorrentFileTreeNode{
			Length:   file.Length,
			Name:     file.Path[len(file.Path)-1],
			Path:     strings.Join(file.Path, "/"),
			Children: &[]*types.TorrentFileTreeNode{},
		})
		if parentNode != nil {
			parentNode.Length += file.Length
		}
	}
	return fileTree
}
