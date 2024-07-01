package db

import "downite/types"

func (db *Database) GetAllTrackers() ([]string, error) {
	var err error
	var trackers []string
	err = db.x.Select(&trackers, `SELECT address FROM trackers`)
	if err != nil {
		return nil, err
	}
	return trackers, err
}

func (db *Database) InsertTracker(tracker *types.Tracker, infohash string) error {
	result, err := db.x.NamedExec(`INSERT INTO trackers (url) VALUES (:url)`, tracker)
	if err != nil {
		return err
	}

	trackerId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	_ = db.x.MustExec(`INSERT INTO torrent_trackers (infohash, tracker_id, tier) VALUES ($1, $2, $3)`, infohash, trackerId, tracker.Tier)
	return nil
}
func (db *Database) GetTorrentTrackers(infohash string) ([]types.Tracker, error) {
	var err error
	var trackers []types.Tracker
	err = db.x.Select(&trackers, `
		SELECT trackers.url, torrent_trackers.tier FROM 
		trackers JOIN torrent_trackers ON torrent_trackers.tracker_id = trackers.id
		WHERE torrent_trackers.infohash = ?`, infohash)
	if err != nil {
		return nil, err
	}
	return trackers, err
}
func (db *Database) DeleteTorrentTrackerLinks(infohash string) error {
	_, err := db.x.Exec(`DELETE FROM torrent_trackers WHERE infohash = ?`, infohash)
	return err
}
