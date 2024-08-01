package db

import (
	"downite/types"
	"fmt"
)

func (db *Database) GetTorrents() ([]types.Torrent, error) {
	var err error
	var torrents []types.Torrent
	err = db.x.Select(&torrents, `
SELECT
	infohash,
	name,
	queue_number,
	save_path,
	status,
	time_active,
	downloaded,
	uploaded,
	total_size,
	size_of_wanted,
	comment,
	category_id,
	created_at,
	started_at
FROM
	torrents
ORDER BY
	created_at DESC
 `)
	if err != nil {
		return nil, err
	}

	return torrents, err
}

func (db *Database) GetTorrent(torrentHash string) (*types.Torrent, error) {
	var err error
	var torrent types.Torrent
	err = db.x.Get(&torrent, `
SELECT
	infohash,
	name,
	queue_number,
	save_path,
	status,
	time_active,
	downloaded,
	uploaded,
	total_size,
	size_of_wanted,
	comment,
	category_id,
	created_at,
	started_at
FROM
	torrents
WHERE
	torrents.infohash = ?
`, torrentHash)
	if err != nil {
		return nil, err
	}
	return &torrent, err
}

func (db *Database) InsertTorrent(torrent *types.Torrent, addTopOfQueue bool) error {
	if addTopOfQueue {
		if torrent.QueueNumber != 1 {
			return fmt.Errorf("cannot add torrent to top of queue with queue number %d", torrent.QueueNumber)
		}
		_, err := db.x.Exec(`UPDATE torrents SET queue_number = queue_number + 1`)
		if err != nil {
			return err
		}
	}
	_, err := db.x.NamedExec(`INSERT INTO torrents
	(created_at, infohash, name, queue_number, save_path, status, time_active, downloaded, uploaded, total_size, size_of_wanted, comment, category_id, created_at, started_at)
	VALUES
	(:created_at, :infohash, :name, :queue_number, :save_path, :status, :time_active, :downloaded, :uploaded, :total_size, :size_of_wanted, :comment, :category_id, :created_at, :started_at)
	`, torrent)
	return err
}

func (db *Database) GetLastQueueNumberOfTorrents() (int, error) {
	var lastQueueNumber int
	err := db.x.Get(&lastQueueNumber, `SELECT MAX(queue_number) FROM torrents`)
	return lastQueueNumber, err
}

func (db *Database) UpdateTorrent(torrent *types.Torrent) error {
	_, err := db.x.NamedExec(`
	UPDATE torrents
	SET
		name = :name,
		queue_number = :queue_number,
		save_path = :save_path,
		status = :status,
		time_active = :time_active,
		downloaded = :downloaded,
		uploaded = :uploaded,
		total_size = :total_size,
		size_of_wanted = :size_of_wanted,
		comment = :comment,
		category_id = :category_id,
		created_at = :created_at,
		started_at = :started_at
	WHERE
		infohash = :infohash
	`, torrent)
	return err
}
func (db *Database) UpdateTorrentStatus(infohash string, status types.TorrentStatus) error {
	_, err := db.x.Exec(`
	UPDATE torrents
	SET
		status = $1 
	WHERE
		infohash = $2 
	`, status.String(), infohash)
	return err
}
func (db *Database) UpdateSizeOfWanted(torrent *types.Torrent) error {
	_, err := db.x.NamedExec(`
	UPDATE torrents
	SET
		size_of_wanted = :size_of_wanted
	WHERE
		infohash = :infohash
	`, torrent)
	return err
}

func (db *Database) DeleteTorrent(torrentHash string) error {
	var queueNumber int
	err := db.x.Get(&queueNumber, "SELECT queue_number FROM torrents WHERE infohash = ?", torrentHash)
	if err != nil {
		return err
	}
	transaction := db.x.MustBegin()
	_, err = transaction.Exec(`DELETE FROM torrents WHERE infohash = ?`, torrentHash)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(`UPDATE torrents SET queue_number = queue_number - 1 WHERE queue_number > ?`, queueNumber)
	if err != nil {
		return err
	}
	err = transaction.Commit()
	return err
}
