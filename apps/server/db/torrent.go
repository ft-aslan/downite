package db

import "downite/types"

func GetTorrents() ([]types.Torrent, error) {
	var torrents []types.Torrent
	err := DB.Select(&torrents, "SELECT * FROM torrents t ORDER BY added_on DESC JOIN torrent_tags tt ON t.infohash = tt.infohash")
	return torrents, err
}
