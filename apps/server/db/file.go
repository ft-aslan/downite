package db

import (
	"downite/types"
)

func InsertTorrentFile(file *types.TorrentFileTreeNode, infohash string) error {
	_, err := DB.Exec(
		`INSERT INTO files (infohash, name, path, priority) VALUES ($1, $2, $3, $4)`,
		infohash, file.Name, file.Path, file.Priority)
	if err != nil {
		return err
	}
	return nil
}
func GetTorrentTorrentFiles(infohash string) ([]types.TorrentFileTreeNode, error) {
	var err error
	var files []types.TorrentFileTreeNode
	err = DB.Select(&files, `SELECT name, path, priority FROM files WHERE infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return files, err
}
func DeleteTorrentFilesByInfohash(infohash string) error {
	_, err := DB.Exec(`DELETE FROM files WHERE infohash = ?`, infohash)
	if err != nil {
		return err
	}
	return nil
}
