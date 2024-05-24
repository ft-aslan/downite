package db

import "downite/types"

func InsertTorrentFile(file *types.TorrentFileInfo, infohash string) error {
	_, err := DB.Exec(
		`INSERT INTO files (infohash, name,path, priority) VALUES ($1, $2, $3, $4)`,
		infohash, file.Name, file.Path, file.Priority)
	if err != nil {
		return err
	}
	return nil
}
func GetTorrentTorrentFiles(infohash string) ([]types.TorrentFileInfo, error) {
	var err error
	var files []types.TorrentFileInfo
	err = DB.Select(&files, `
		SELECT files.url, torrent_trackers.tier FROM 
		files JOIN torrent_trackers ON torrent_trackers.tracker_id = trackers.id
		WHERE torrent_files.infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return files, err
}
