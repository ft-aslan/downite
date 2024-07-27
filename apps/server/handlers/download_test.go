package handlers_test

import (
	"downite/api"
	"downite/db"
	"downite/handlers"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
)

func TestDownload(t *testing.T) {
	_, humaTestApi := humatest.New(t)

	db, err := db.DbInit()
	if err != nil {
		t.Errorf("Cannot connect to db : %s", err)
	}
	engine, err := api.InitDownloadEngine(db)
	if err != nil {
		t.Errorf("Cannot initilize download engine : %s", err)
	}
	api.AddDownloadRoutes(handlers.DownloadHandler{
		Engine: engine,
		Db:     db,
	}, humaTestApi)
	res := humaTestApi.Post("/api/download", map[string]any{
		"url":                         "https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso",
		"category":                    "",
		"savePath":                    "",
		"isIncompleteSavePathEnabled": false,
		"incompleteSavePath":          "",
		"contentLayout":               "Original",
		"tags":                        []string{},
		"startDownload":               true,
		"addTopOfQueue":               false,
	})
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
}
