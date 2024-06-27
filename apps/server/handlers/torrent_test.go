package handlers_test

import (
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
)

const testTorrentFile = "./mocks/debian-12.5.0-amd64-netinst.iso.torrent"
const testMagnetLink = "magnet:?xt=urn:btih:2b66980093bc11806fab50cb3cb41835b95a0362&dn=debian-12.5.0-amd64-netinst.iso&tr=http%3A%2F%2Fbttracker.debian.org%3A6969%2Fannounce&ws=https%3A%2F%2Fcdimage.debian.org%2Fcdimage%2Frelease%2F12.5.0%2Famd64%2Fiso-cd%2Fdebian-12.5.0-amd64-netinst.iso&ws=https%3A%2F%2Fcdimage.debian.org%2Fcdimage%2Farchive%2F12.5.0%2Famd64%2Fiso-cd%2Fdebian-12.5.0-amd64-netinst.iso"

func TestTorrentDownload(t *testing.T) {
	_, api := humatest.New(t)

}
