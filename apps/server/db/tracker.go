package db

import "downite/types"

func GetAllTrackers() ([]string, error) {
	var err error
	var trackers []string
	err = DB.Select(&trackers, `SELECT address FROM trackers`)
	if err != nil {
		return nil, err
	}
	return trackers, err
}

func AddTracker(tracker types.Tracker, infohash string) error {
	result := DB.MustExec(`INSERT INTO trackers (url) VALUES (?)`, tracker)

	trackerId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	DB.MustExec(`INSERT INTO torrent_trackers (infohash, tracker_id, tier) VALUES (?, ?)`, infohash, trackerId, tracker.Tier)
	return nil
}
func GetTorrentTrackers(infohash string) ([]types.Tracker, error) {
	var err error
	var trackers []types.Tracker
	err = DB.Select(&trackers, `
		SELECT trackers.url, torrent_trackers.tier FROM 
		trackers JOIN torrent_trackers ON torrent_trackers.tracker_id = trackers.id
		WHERE torrent_trackers.infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return trackers, err
}
