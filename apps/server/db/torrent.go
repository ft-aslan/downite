package db

import "downite/types"

func GetTorrents() ([]types.Torrent, error) {
	var err error
	var torrents []types.Torrent
	err = DB.Select(&torrents, `
SELECT
	torrents.id,
	torrents.infohash,
	torrents.name,
	torrents.save_path,
	torrents.status,
	torrents.time_active,
	torrents.downloaded,
	torrents.uploaded,
	torrents.total_size,
	torrents.comment,
	torrents.category_id,
	torrents.created_at,
	torrents.started_at
FROM
	torrents
ORDER BY
	added_on DESC
 `)
	if err != nil {
		return nil, err
	}

	return torrents, err
}

func GetTorrent(torrentHash string) (*types.Torrent, error) {
	var err error
	var torrent types.Torrent
	err = DB.Get(&torrent, `
SELECT
	torrents.id,
	torrents.infohash,
	torrents.name,
	torrents.save_path,
	torrents.status,
	torrents.time_active,
	torrents.downloaded,
	torrents.uploaded,
	torrents.total_size,
	torrents.comment,
	torrents.category_id,
	torrents.created_at,
	torrents.started_at
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

func InsertTorrent(torrent *types.Torrent) error {
	_, err := DB.NamedExec(`INSERT INTO torrents
	(infohash, name, save_path, status, time_active, downloaded, uploaded, total_size, comment, category_id, created_at, started_at)
	VALUES
	(:infohash, :name, :save_path, :status, :time_active, :downloaded, :uploaded, :total_size, :comment, :category_id, :created_at, :started_at)
	`, torrent)
	return err
}

func UpdateTorrent(torrent *types.Torrent) error {
	_, err := DB.NamedExec(`
	UPDATE torrents
	SET
		name = :name,
		save_path = :save_path,
		status = :status,
		time_active = :time_active,
		downloaded = :downloaded,
		uploaded = :uploaded,
		total_size = :total_size,
		comment = :comment,
		category_id = :category_id,
		created_at = :created_at,
		started_at = :started_at
	WHERE
		infohash = :infohash
	`, torrent)
	return err
}
func UpdateTorrentStatus(torrent *types.Torrent) error {
	_, err := DB.NamedExec(`
	UPDATE torrents
	SET
		status = :status
	WHERE
		infohash = :infohash
	`, torrent)
	return err
}
