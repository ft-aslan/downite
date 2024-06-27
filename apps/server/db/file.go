package db

import (
	"downite/types"
)

func (db *Database) InsertTorrentFile(file *types.TorrentFileTreeNode, infohash string) error {
	_, err := db.x.Exec(
		`INSERT INTO files (infohash, name, path, priority) VALUES ($1, $2, $3, $4)`,
		infohash, file.Name, file.Path, file.Priority)
	if err != nil {
		return err
	}
	return nil
}
func (db *Database) GetTorrentTorrentFiles(infohash string) ([]types.TorrentFileTreeNode, error) {
	var err error
	var files []types.TorrentFileTreeNode
	err = db.x.Select(&files, `SELECT name, path, priority FROM files WHERE infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return files, err
}
func (db *Database) DeleteTorrentFilesByInfohash(infohash string) error {
	_, err := db.x.Exec(`DELETE FROM files WHERE infohash = ?`, infohash)
	if err != nil {
		return err
	}
	return nil
}
