package db

import "downite/types"

func (db *Database) InsertDownload(download *types.Download) error {
	_, err := db.x.NamedExec(`INSERT INTO downloads
	(status, name, path, part_count, part_length, total_size, downloaded_bytes, url, queue_number)
	VALUES
	(:status, :name, :path, :part_count, :part_length, :total_size, :downloaded_bytes, :url, :queue_number)
	`, download)
	return err
}
