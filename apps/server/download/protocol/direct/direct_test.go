package direct_test

import (
	"downite/db"
	"downite/download/protocol/direct"
	"downite/types"
	"testing"
)

func initDownloadTest(t *testing.T) *direct.Client {
	//initilize db
	db, err := db.DbInit()
	if err != nil {
		t.Errorf("Cannot connect to db : %s", err)
	}
	defaultClientConfig, err := direct.NewClientDefaultConfig()
	if err != nil {
		t.Errorf("Cannot get default config : %s", err)
	}
	//initilize download client
	client, err := direct.CreateDownloadClient(defaultClientConfig, db)
	if err != nil {
		t.Errorf("Cannot create download client : %s", err)
	}
	err = client.InitDownloads()
	if err != nil {
		t.Errorf("Cannot initilize downloads : %s", err)
	}

	return client
}
func TestDownloadFromUrl(t *testing.T) {
	client := initDownloadTest(t)
	_, err := client.DownloadFromUrl("https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso", 8, "", true, false)
	if err != nil {
		t.Errorf("Cannot create download : %s", err)
	}
}
func pauseScenarioInit(t *testing.T, status types.DownloadStatus) {
	client := initDownloadTest(t)
	var download *types.Download
	var err error
	if status == types.DownloadStatusDownloading {
		download, err = client.DownloadFromUrl("https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso", 8, "", true, false)
		if err != nil {
			t.Errorf("Cannot create download : %s", err)
		}
	} else if status == types.DownloadStatusPaused {
		download, err = client.DownloadFromUrl("https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso", 8, "", false, false)
		if err != nil {
			t.Errorf("Cannot create download : %s", err)
		}
	}

	err = client.PauseDownload(download.Id)
	if err != nil {
		if status == types.DownloadStatusPaused {
			if err.Error() == "download is already paused" {
				// its ok
				return
			}
		}

		t.Errorf("Cannot pause download : %s", err)
	}

}

func TestPauseDownloadIfStateDownloading(t *testing.T) {
	pauseScenarioInit(t, types.DownloadStatusDownloading)
}

func TestPauseDownloadIfStatePaused(t *testing.T) {
	pauseScenarioInit(t, types.DownloadStatusPaused)
}
