package handlers

import (
	"context"
	"downite/db"
	"downite/download/protocol/http"
)

type DownloadHandler struct {
	Db     *db.Database
	Client *http.Client
}

type GetDownloadFileInfoReq struct {
	Body struct {
		Url string `json:"url" minLength:"1"`
	}
}
type GetDownloadFileInfoRes struct {
	Body http.Download
}

func (downloadHandler *DownloadHandler) GetDownloadFileInfo(ctx context.Context, input *GetDownloadFileInfoReq) (*GetDownloadFileInfoRes, error) {
	res := &GetDownloadFileInfoRes{}
	return res, nil
}

type DownloadReq struct {
	Body struct {
		Url         string   `json:"url" minLength:"1"`
		Category    string   `json:"category"`
		Path        string   `json:"path"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
	}
}
type DownloadRes struct {
	Body http.Download
}

func (downloadHandler *DownloadHandler) Download(ctx context.Context, input *DownloadReq) (*DownloadRes, error) {
	res := &DownloadRes{}

	return res, nil
}

type GetDownloadReq struct {
	id string `path:"id"`
}
type GetDownloadRes struct {
	Body http.Download
}

func (downloadHandler *DownloadHandler) GetDownload(ctx context.Context, input *GetDownloadReq) (*GetDownloadRes, error) {
	res := &GetDownloadRes{}

	return res, nil
}
