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

	gotorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	gotorrenttypes "github.com/anacrolix/torrent/types"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/danielgtaylor/huma/v2"
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
		dbTorrent, err := torr.GetTorrentDetails(torrent)
		if err != nil {
			return nil, err
		}
		torrentsRes = append(torrentsRes, *dbTorrent)
	}

	res.Body.Torrents = torrentsRes
	return res, nil
}

type GetTorrentReq struct {
	Infohash string `path:"infohash" maxLength:"40" example:"2b66980093bc11806fab50cb3cb41835b95a0362" doc:"Infohash of the torrent"`
}
type GetTorrentRes struct {
	Body types.Torrent
}

func GetTorrent(ctx context.Context, input *GetTorrentReq) (*GetTorrentRes, error) {
	res := &GetTorrentRes{}
	torrent, ok := torr.Client.Torrent(infohash.FromHexString(input.Infohash))
	if !ok {
		return nil, fmt.Errorf("torrent with hash %s not found", input.Infohash)
	}
	dbTorrent, err := torr.GetTorrentDetails(torrent)
	if err != nil {
		return nil, err
	}

	res.Body = *dbTorrent

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
			foundTorrent.SetMaxEstablishedConns(0)
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
			foundTorrent.SetMaxEstablishedConns(80)
			torr.StartTorrent(foundTorrent)

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

	// Insert torrent
	err = db.InsertTorrent(&dbTorrent)
	if err != nil {
		return nil, err
	}

	// Insert trackers
	for _, dbTracker := range dbTorrent.Trackers {
		if err = db.InsertTracker(&dbTracker, dbTorrent.Infohash); err != nil {
			return nil, err
		}
	}

	// ADD TORRENT TO CLIENT
	torrent, err = torr.AddTorrent(dbTorrent.Infohash, dbTorrent.Trackers, dbTorrent.SavePath, !input.Body.SkipHashCheck)
	if err != nil {
		return nil, err
	}

	// Insert download priorities of the files
	for _, file := range torrent.Files() {
		for _, clientFile := range input.Body.Files {
			if file.DisplayPath() == clientFile.Path {
				priority, ok := types.PiecePriorityStringMap[clientFile.Priority]
				if !ok {
					return nil, fmt.Errorf("invalid download priority: %s", clientFile.Priority)
				}

				if priority != gotorrenttypes.PiecePriorityNone {
					dbTorrent.SizeOfWanted += file.Length()
				}

				var fileName string
				//if its not multi file torrentt path array gonna be empty. use display path instead
				if len(file.FileInfo().Path) == 0 {
					fileName = file.DisplayPath()
				} else {
					fileName = file.FileInfo().Path[len(file.FileInfo().Path)-1]
				}
				db.InsertTorrentFile(&types.TorrentFileTreeNode{
					Path:     file.Path(),
					Priority: clientFile.Priority,
					Name:     fileName,
				}, dbTorrent.Infohash)
			}
		}

	}

	dbTorrent.Files = torr.CreateFileTreeFromMeta(*torrent.Info())

	if input.Body.StartTorrent {
		torrent, err = torr.StartTorrent(torrent)
		if err != nil {
			return nil, err
		}
		dbTorrent.Status = types.TorrentStatusStringMap[types.TorrentStatusDownloading]
	} else {
		dbTorrent.Status = types.TorrentStatusStringMap[types.TorrentStatusPaused]
	}

	dbTorrent.TotalSize = torrent.Length()
	dbTorrent.Magnet = torrent.Metainfo().Magnet(nil, torrent.Info()).String()

	db.UpdateTorrent(&dbTorrent)

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

	fileTree := torr.CreateFileTreeFromMeta(info)

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
		ctx.Err()
		return nil, huma.Error400BadRequest("invalid magnet")
	}
	// Load from a magnet link

	torrent, err := torr.Client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	<-torrent.GotInfo()

	info = *torrent.Info()
	infohash = torrent.InfoHash().String()

	fileTree := torr.CreateFileTreeFromMeta(info)

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
