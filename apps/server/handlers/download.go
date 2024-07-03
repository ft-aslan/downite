package handlers

import (
	"context"
	"downite/db"
	"downite/download/protocol/http"
	"downite/types"
)

type DownloadHandler struct {
	Db     *db.Database
	Engine *http.Client
}

type GetDownloadFileInfoReq struct {
	Body struct {
		Url string `json:"url" minLength:"1"`
	}
}
type GetDownloadFileInfoRes struct {
	Body types.Download
}

func (handler *DownloadHandler) GetDownloadFileInfo(ctx context.Context, input *GetDownloadFileInfoReq) (*GetDownloadFileInfoRes, error) {
	res := &GetDownloadFileInfoRes{}
	return res, nil
}

type DownloadReq struct {
	Body struct {
		Url         string   `json:"url" minLength:"1" uri:"true"`
		Category    string   `json:"category"`
		Path        string   `json:"path"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
	}
}
type DownloadRes struct {
	Body types.Download
}

func (handler *DownloadHandler) Download(ctx context.Context, input *DownloadReq) (*DownloadRes, error) {
	res := &DownloadRes{}
	err := handler.Engine.DownloadFromUrl(input.Body.Url, handler.Engine.Config.PartCount, input.Body.Path)
	return res, err
}

type GetDownloadReq struct {
	id string `path:"id"`
}
type GetDownloadRes struct {
	Body types.Download
}

func (handler *DownloadHandler) GetDownload(ctx context.Context, input *GetDownloadReq) (*GetDownloadRes, error) {
	res := &GetDownloadRes{}

	return res, nil
}
