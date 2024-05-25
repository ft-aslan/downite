package db

import "downite/types"

func InsertTorrentFile(file *types.TorrentFile, infohash string) error {
	_, err := DB.Exec(
		`INSERT INTO files (infohash, name, path, priority) VALUES ($1, $2, $3, $4)`,
		infohash, file.Name, file.Path, file.Priority)
	if err != nil {
		return err
	}
	return nil
}
func GetTorrentTorrentFiles(infohash string) ([]types.TorrentFile, error) {
	var err error
	var files []types.TorrentFile
	err = DB.Select(&files, `SELECT name, path, priority FROM files WHERE infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return files, err
}
