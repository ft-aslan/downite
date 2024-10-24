package handlers

import (
	"context"
	"downite/db"
	"downite/download/protocol/direct"
	"downite/types"
	"sort"
	"strconv"
	"time"
)

type DownloadHandler struct {
	Db     *db.Database
	Engine *direct.DirectDownloadEngine
}

type DownloadsTotalSpeedData struct {
	DownloadSpeed uint64 `json:"downloadSpeed"`
	Time          string `json:"time"`
}
type GetDownloadsTotalSpeedRes struct {
	Body DownloadsTotalSpeedData
}

func (handler *DownloadHandler) GetDownloadsTotalSpeed(ctx context.Context, input *struct{}) (*GetDownloadsTotalSpeedRes, error) {
	res := &GetDownloadsTotalSpeedRes{}
	res.Body.DownloadSpeed = handler.Engine.GetTotalDownloadSpeed()
	res.Body.Time = time.Now().Format("15:04:05")
	return res, nil
}

type GetDownloadMetaReq struct {
	Body struct {
		Url string `json:"url" minLength:"1"`
	}
}
type GetDownloadMetaRes struct {
	Body types.DownloadMeta
}

func (handler *DownloadHandler) GetDownloadMeta(ctx context.Context, input *GetDownloadMetaReq) (*GetDownloadMetaRes, error) {
	res := &GetDownloadMetaRes{}

	metaInfo, err := handler.Engine.GetDownloadMeta(input.Body.Url)
	if err != nil {
		return res, err
	}
	res.Body = *metaInfo
	return res, nil
}

type DownloadReq struct {
	Body struct {
		Name                        string   `json:"name"`
		Url                         string   `json:"url" minLength:"1" uri:"true"`
		Category                    string   `json:"category"`
		SavePath                    string   `json:"savePath"`
		IsIncompleteSavePathEnabled bool     `json:"isIncompleteSavePathEnabled"`
		IncompleteSavePath          string   `json:"incompleteSavePath"`
		ContentLayout               string   `json:"contentLayout" enum:"Original,Create subfolder,Don't create subfolder"`
		Tags                        []string `json:"tags"`
		StartDownload               bool     `json:"startDownload"`
		AddTopOfQueue               bool     `json:"addTopOfQueue"`
		Overwrite                   bool     `json:"overwrite"`
	}
}
type DownloadRes struct {
	Body *types.Download
}

func (handler *DownloadHandler) Download(ctx context.Context, input *DownloadReq) (*DownloadRes, error) {
	res := &DownloadRes{}
	download, err := handler.Engine.DownloadFromUrl(input.Body.Name, input.Body.Url, handler.Engine.DownloadClientConfig.PartCount, input.Body.SavePath, input.Body.StartDownload, input.Body.AddTopOfQueue, input.Body.Overwrite)
	if err != nil {
		return nil, err
	}
	res.Body = download
	return res, err
}

type GetDownloadsRes struct {
	Body []*types.Download
}

func (handler *DownloadHandler) GetDownloads(ctx context.Context, input *struct{}) (*GetDownloadsRes, error) {
	res := &GetDownloadsRes{}
	downloads, err := handler.Engine.GetDownloads()
	sort.Slice(downloads, func(i, j int) bool {
		return downloads[i].QueueNumber < downloads[j].QueueNumber
	})
	if err != nil {
		return nil, err
	}
	res.Body = downloads
	return res, nil
}

type GetDownloadReq struct {
	Id string `path:"id"`
}
type GetDownloadRes struct {
	Body *types.Download
}

func (handler *DownloadHandler) GetDownload(ctx context.Context, input *GetDownloadReq) (*GetDownloadRes, error) {
	res := &GetDownloadRes{}
	id, _ := strconv.Atoi(input.Id)
	download, err := handler.Engine.GetDownload(id)
	if err != nil {
		return nil, err
	}
	res.Body = download

	return res, nil
}

type GetNewFileNameForPathReq struct {
	Body struct {
		SavePath string `json:"savePath"`
		FileName string `json:"fileName"`
	}
}
type GetNewFileNameForPathRes struct {
	Body string
}

func (handler *DownloadHandler) GetNewFileNameForPath(ctx context.Context, input *GetNewFileNameForPathReq) (*GetNewFileNameForPathRes, error) {
	res := &GetNewFileNameForPathRes{}
	newFileName, err := handler.Engine.CreateNewFileNameForPath(input.Body.SavePath, input.Body.FileName)
	if err != nil {
		return nil, err
	}
	res.Body = newFileName
	return res, nil
}

type DownloadActionReq struct {
	Body struct {
		Ids []int `json:"ids"`
	}
}
type DownloadActionRes struct {
	Body struct{}
}

func (handler *DownloadHandler) PauseDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	for _, id := range input.Body.Ids {
		err := handler.Engine.PauseDownload(id)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (handler *DownloadHandler) ResumeDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	for _, id := range input.Body.Ids {
		err := handler.Engine.ResumeDownload(id)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (handler *DownloadHandler) DeleteDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	for _, id := range input.Body.Ids {
		err := handler.Engine.DeleteDownload(id)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (handler *DownloadHandler) RemoveDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	for _, id := range input.Body.Ids {
		err := handler.Engine.RemoveDownload(id)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
