package db

func GetAllTrackers() ([]string, error) {
	var err error
	var trackers []string
	err = DB.Select(&trackers, `SELECT address FROM trackers`)
	if err != nil {
		return nil, err
	}
	return trackers, err
}

func AddTracker(address string, infohash string) error {
	result := DB.MustExec(`INSERT INTO trackers (address) VALUES (?)`, address)

	trackerId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	DB.MustExec(`INSERT INTO torrent_trackers (infohash, tracker_id) VALUES (?, ?)`, infohash, trackerId)
	return nil
}
func GetTorrentTrackers(infohash string) ([]string, error) {
	var err error
	var trackers []string
	err = DB.Select(&trackers, `
		SELECT trackers.address FROM trackers
		JOIN torrent_trackers ON torrent_trackers.tracker_id = trackers.id
		WHERE torrent_trackers.infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return trackers, err
}
