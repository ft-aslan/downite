package handlers

import (
	"context"
	"downite/settings"
)

type SettingsHandler struct {
	SettingsSystem *settings.DowniteSettingsSystem
}

type AddSavePathReq struct {
	Body string
}
type AddSavePathRes struct {
	Body bool
}

func (handler *SettingsHandler) AddSavePath(ctx context.Context, input *AddSavePathReq) (*AddSavePathRes, error) {
	res := &AddSavePathRes{}
	err := handler.SettingsSystem.AddSavePath(input.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type GetSavePathsRes struct {
	Body []string
}

func (handler *SettingsHandler) GetSavePaths(ctx context.Context, input *struct{}) (*GetSavePathsRes, error) {
	res := &GetSavePathsRes{}
	res.Body = handler.SettingsSystem.Settings.SavePaths
	return res, nil
}
