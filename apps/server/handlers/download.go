package handlers

import (
	"context"
	"downite/db"
	"downite/download/protocol/direct"
	"downite/types"
)

type DownloadHandler struct {
	Db     *db.Database
	Engine *direct.Client
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
		Url                         string   `json:"url" minLength:"1" uri:"true"`
		Category                    string   `json:"category"`
		SavePath                    string   `json:"savePath"`
		IsIncompleteSavePathEnabled bool     `json:"isIncompleteSavePathEnabled"`
		IncompleteSavePath          string   `json:"incompleteSavePath"`
		ContentLayout               string   `json:"contentLayout" enum:"Original,Create subfolder,Don't create subfolder"`
		Tags                        []string `json:"tags"`
		Description                 string   `json:"description"`
		StartDownload               bool     `json:"startDownload"`
		AddTopOfQueue               bool     `json:"addTopOfQueue"`
	}
}
type DownloadRes struct {
	Body *types.Download
}

func (handler *DownloadHandler) Download(ctx context.Context, input *DownloadReq) (*DownloadRes, error) {
	res := &DownloadRes{}
	download, err := handler.Engine.DownloadFromUrl(input.Body.Url, handler.Engine.Config.PartCount, input.Body.SavePath, input.Body.StartDownload)
	res.Body = download
	return res, err
}

type GetDownloadsRes struct {
	Body []*types.Download
}

func (handler *DownloadHandler) GetDownloads(ctx context.Context, input *struct{}) (*GetDownloadsRes, error) {
	res := &GetDownloadsRes{}
	downloads, err := handler.Engine.GetDownloads()
	if err != nil {
		return nil, err
	}
	res.Body = downloads
	return res, nil
}

type GetDownloadReq struct {
	id string `path:"id"`
}
type GetDownloadRes struct {
	Body *types.Download
}

func (handler *DownloadHandler) GetDownload(ctx context.Context, input *GetDownloadReq) (*GetDownloadRes, error) {
	res := &GetDownloadRes{}

	return res, nil
}

type DownloadActionReq struct {
	Body struct {
		Ids []int `json:"ids"`
	}
}
type DownloadActionRes struct {
	Body struct {
		Success bool `json:"success"`
	}
}

func (handler *DownloadHandler) PauseDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	return res, nil
}
func (handler *DownloadHandler) ResumeDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	return res, nil
}
func (handler *DownloadHandler) DeleteDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	return res, nil
}
func (handler *DownloadHandler) RemoveDownload(ctx context.Context, input *DownloadActionReq) (*DownloadActionRes, error) {
	res := &DownloadActionRes{}
	return res, nil
}
