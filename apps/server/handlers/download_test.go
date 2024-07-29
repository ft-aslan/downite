package handlers_test

import (
	"downite/api"
	"downite/db"
	"downite/handlers"
	"downite/types"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
)

var downloadReqMock = map[string]any{
	"url":                         "https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso",
	"category":                    "",
	"savePath":                    "",
	"isIncompleteSavePathEnabled": false,
	"incompleteSavePath":          "",
	"contentLayout":               "Original",
	"tags":                        []string{},
	"startDownload":               false,
	"addTopOfQueue":               false,
}

func initDownloadTest(t *testing.T) humatest.TestAPI {
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
	return humaTestApi
}

func TestDownload(t *testing.T) {
	testApi := initDownloadTest(t)
	res := testApi.Post("/download", downloadReqMock)
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
}

func TestGetDownload(t *testing.T) {
	testApi := initDownloadTest(t)
	res := testApi.Post("/download", downloadReqMock)
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
	downloadRes := types.Download{}
	err := json.Unmarshal(res.Body.Bytes(), &downloadRes)
	if err != nil {
		t.Errorf("cannot unmarshal response : %s", err)
	}
	res = testApi.Get(fmt.Sprintf("/download/%d", downloadRes.Id))
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
}
func TestDownloadPause(t *testing.T) {
	testApi := initDownloadTest(t)

	downloadReqMock["startDownload"] = true
	res := testApi.Post("/download", downloadReqMock)
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
	downloadRes := types.Download{}
	err := json.Unmarshal(res.Body.Bytes(), &downloadRes)
	if err != nil {
		t.Errorf("cannot unmarshal response : %s", err)
	}
	req := map[string]any{
		"ids": []int{downloadRes.Id},
	}
	res = testApi.Post("/download/pause", req)
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}

	res = testApi.Post("/download/delete", req)
	if res.Code != 200 {
		t.Errorf("Expected code 200 got code %d and response %v", res.Code, res.Body)
	}
}
