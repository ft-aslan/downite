package db

import "downite/types"

func (db *Database) InsertDownload(download *types.Download) (int, error) {
	result, err := db.x.NamedExec(`INSERT INTO downloads
	(status, name, path, part_count, part_length, total_size, downloaded_bytes, url, queue_number)
	VALUES
	(:status, :name, :path, :part_count, :part_length, :total_size, :downloaded_bytes, :url, :queue_number)
	`, download)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), err
}
