package db

import (
	"downite/types"
	"fmt"
)

func (db *Database) InsertDownload(download *types.Download, addTopOfQueue bool) (int, error) {
	if addTopOfQueue {
		if download.QueueNumber != 1 {
			return 0, fmt.Errorf("cannot add download to top of queue with queue number %d", download.QueueNumber)
		}
		_, err := db.x.Exec(`UPDATE downloads SET queue_number = queue_number + 1`)
		if err != nil {
			return 0, err
		}
	}
	result, err := db.x.NamedExec(`INSERT INTO downloads
	(created_at, status, name, save_path, part_count, part_length, total_size, downloaded_bytes, url, queue_number, error)
	VALUES
	(:created_at, :status, :name, :save_path, :part_count, :part_length, :total_size, :downloaded_bytes, :url, :queue_number, :error)
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
func (db *Database) GetDownload(id int) (*types.Download, error) {
	var err error
	var download types.Download
	err = db.x.Get(&download, `SELECT * FROM downloads WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &download, err
}

func (db *Database) GetDownloads() ([]types.Download, error) {
	var err error
	var downloads []types.Download
	err = db.x.Select(&downloads, `SELECT * FROM downloads`)
	if err != nil {
		return nil, err
	}
	return downloads, err
}

func (db *Database) DeleteDownload(id int) error {
	_, err := db.x.Exec(`DELETE FROM downloads WHERE id = ?`, id)
	return err
}

func (db *Database) DeleteDownloadParts(id int) error {
	_, err := db.x.Exec(`DELETE FROM download_parts WHERE download_id = ?`, id)
	return err
}
func (db *Database) GetLastQueueNumberOfDownloads() (int, error) {
	var lastQueueNumber int
	err := db.x.Get(&lastQueueNumber, `SELECT MAX(queue_number) FROM downloads`)
	return lastQueueNumber, err
}

func (db *Database) UpdateDownload(download *types.Download) error {
	_, err := db.x.NamedExec(`UPDATE downloads
	SET
		started_at = :started_at,
		time_active = :time_active,
		finished_at = :finished_at,
		status = :status,
		name = :name,
		save_path = :save_path,
		part_count = :part_count,
		part_length = :part_length,
		total_size = :total_size,
		downloaded_bytes = :downloaded_bytes,
		url = :url,
		queue_number = :queue_number,
		error = :error
	WHERE
		id = :id
	`, download)
	return err
}
func (db *Database) InsertDownloadParts(downloadPart []*types.DownloadPart) error {
	_, err := db.x.NamedExec(`INSERT INTO download_parts
	(created_at, status, part_index, start_byte_index, end_byte_index, part_length, downloaded_bytes, download_id)
	VALUES
	(:created_at, :status, :part_index, :start_byte_index, :end_byte_index, :part_length, :downloaded_bytes, :download_id)
	`, downloadPart)

	return err
}
func (db *Database) GetDownloadParts(downloadId int) ([]*types.DownloadPart, error) {
	var err error
	var downloadParts []*types.DownloadPart
	err = db.x.Select(&downloadParts, `SELECT * FROM download_parts WHERE download_id = ?`, downloadId)
	if err != nil {
		return nil, err
	}
	return downloadParts, err
}
func (db *Database) UpdateDownloadPart(downloadPart *types.DownloadPart) error {
	_, err := db.x.NamedExec(`UPDATE download_parts
	SET
		started_at = :started_at,
		time_active = :time_active,
		finished_at = :finished_at,
		status = :status,
		downloaded_bytes = :downloaded_bytes
	WHERE
		download_id = :download_id
		AND part_index = :part_index	
	`, downloadPart)
	return err
}
