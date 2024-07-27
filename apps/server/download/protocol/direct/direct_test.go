package direct_test

import (
	"downite/db"
	"downite/download/protocol/direct"
	"testing"
)

func TestCreateDownloadClient(t *testing.T) {
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

	if client == nil {
		t.Errorf("download client is nil")
	}
}
