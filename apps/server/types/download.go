package types

import "time"

type DownloadStatus int

const (
	DownloadStatusPaused DownloadStatus = iota
	DownloadStatusDownloading
	DownloadStatusCompleted
	DownloadStatusError
)

var DownloadStatusStringMap = map[DownloadStatus]string{
	DownloadStatusPaused:      "paused",
	DownloadStatusDownloading: "downloading",
	DownloadStatusCompleted:   "completed",
	DownloadStatusError:       "error",
}

func (d DownloadStatus) String() string {
	return DownloadStatusStringMap[d]
}

type Download struct {
	Id              int           `json:"id"`
	CreatedAt       time.Time     `json:"createdAt" db:"created_at"`
	StartedAt       time.Time     `json:"startedAt" db:"started_at"`
	TimeActive      time.Duration `json:"timeActive" db:"time_active"`
	FinishedAt      time.Time     `json:"finishedAt" db:"finished_at"`
	Status          DownloadStatus
	Name            string
	Path            string
	PartCount       int    `db:"part_count"`
	PartLength      uint64 `db:"part_length"`
	TotalSize       uint64 `db:"total_size"`
	DownloadedBytes uint64 `db:"downloaded_bytes"`
	PartProgress    []*DownloadPart
	Url             string
	QueueNumber     int `db:"queue_number"`
}
type DownloadPart struct {
	CreatedAt       time.Time     `db:"created_at"`
	StartedAt       time.Time     `db:"started_at"`
	TimeActive      time.Duration `db:"time_active"`
	FinishedAt      time.Time     `db:"finished_at"`
	Status          DownloadStatus
	PartIndex       int    `db:"part_index"`
	StartByteIndex  uint64 `db:"start_byte_index"`
	EndByteIndex    uint64 `db:"end_byte_index"`
	Buffer          []byte `db:"-"`
	DownloadedBytes uint64 `db:"downloaded_bytes"`
}
