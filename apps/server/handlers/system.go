package handlers

import (
	"context"
	"downite/system"
)

type SystemHandler struct {
	Engine *system.SystemEngine
}
type GetFileSystemNodesReq struct {
	Body struct {
		Path string `json:"path" uri-reference:"true"`
	}
}

type GetFileSystemNodesRes struct {
	Body struct {
		FileSystemNodes []system.FileSystemNode `json:"fileSystemNodes"`
	}
}

func (handler *SystemHandler) GetFileSystemNodes(ctx context.Context, input *GetFileSystemNodesReq) (*GetFileSystemNodesRes, error) {
	res := &GetFileSystemNodesRes{}
	fileSystemNodes, err := handler.Engine.GetFileSystemNodes(input.Body.Path)
	if err != nil {
		return nil, err
	}
	res.Body.FileSystemNodes = fileSystemNodes
	return res, nil
}
