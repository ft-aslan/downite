package handlers_test

import (
	"testing"

	downiteapi "downite/api"
	"downite/handlers"
	"downite/types"

	"github.com/danielgtaylor/huma/v2/humatest"
)

const testTorrentFile = "./mocks/debian-12.5.0-amd64-netinst.iso.torrent"
const testMagnetLink = "magnet:?xt=urn:btih:2b66980093bc11806fab50cb3cb41835b95a0362&dn=debian-12.5.0-amd64-netinst.iso&tr=http%3A%2F%2Fbttracker.debian.org%3A6969%2Fannounce&ws=https%3A%2F%2Fcdimage.debian.org%2Fcdimage%2Frelease%2F12.5.0%2Famd64%2Fiso-cd%2Fdebian-12.5.0-amd64-netinst.iso&ws=https%3A%2F%2Fcdimage.debian.org%2Fcdimage%2Farchive%2F12.5.0%2Famd64%2Fiso-cd%2Fdebian-12.5.0-amd64-netinst.iso"

func TestTorrentDownload(t *testing.T) {
	_, api := humatest.New(t)

	downiteapi.AddRoutes(api)

	req := handlers.DownloadTorrentReq{}
	req.Body.Magnet = testMagnetLink
	req.Body.TorrentFile = ""
	req.Body.SavePath = ""
	req.Body.IsIncompleteSavePathEnabled = false
	req.Body.IncompleteSavePath = ""
	req.Body.Category = ""
	req.Body.Tags = []string{}
	req.Body.StartTorrent = false
	req.Body.AddTopOfQueue = false
	req.Body.DownloadSequentially = false
	req.Body.SkipHashCheck = false
	req.Body.ContentLayout = ""
	req.Body.Files = []types.TorrentFileInfo{}

	res := api.Post("/torrent", req.Body)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}
