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
	Id              int             `json:"id"`
	CreatedAt       time.Time       `json:"createdAt" db:"created_at"`
	StartedAt       time.Time       `json:"startedAt" db:"started_at"`
	TimeActive      time.Duration   `json:"timeActive" db:"time_active"`
	FinishedAt      time.Time       `json:"finishedAt" db:"finished_at"`
	Status          DownloadStatus  `json:"status"`
	Name            string          `json:"name"`
	SavePath        string          `db:"save_path" json:"savePath"`
	PartCount       int             `db:"part_count"`
	PartLength      uint64          `db:"part_length"`
	TotalSize       uint64          `db:"total_size"`
	DownloadedBytes uint64          `db:"downloaded_bytes"`
	Progress        float64         `json:"progress"`
	Parts           []*DownloadPart `json:"parts"`
	Url             string          `json:"url"`
	QueueNumber     int             `db:"queue_number"`
}
type DownloadPart struct {
	CreatedAt       time.Time      `db:"created_at" json:"createdAt"`
	StartedAt       time.Time      `db:"started_at" json:"startedAt"`
	TimeActive      time.Duration  `db:"time_active" json:"timeActive"`
	FinishedAt      time.Time      `db:"finished_at" json:"finishedAt"`
	Status          DownloadStatus `json:"status"`
	PartIndex       int            `db:"part_index" json:"partIndex"`
	StartByteIndex  uint64         `db:"start_byte_index" json:"startByteIndex"`
	EndByteIndex    uint64         `db:"end_byte_index" json:"endByteIndex"`
	PartLength      uint64         `db:"part_length" json:"partLength"`
	DownloadedBytes uint64         `db:"downloaded_bytes" json:"downloadedBytes"`
	Progress        float64        `db:"-" json:"progress"`
	//we dont store buffer in db and memory
	Buffer     []byte `db:"-" json:"-"`
	DownloadId int    `json:"-" db:"download_id"`
}
type DownloadMeta struct {
	TotalSize      uint64 `json:"totalSize"`
	Url            string `json:"url"`
	FileName       string `json:"fileName"`
	FileType       string `json:"fileType"`
	IsRangeAllowed bool   `json:"isRangeAllowed"`
}

func (d *DownloadPart) Write(bytes []byte) (int, error) {
	d.DownloadedBytes += uint64(len(bytes))
	return len(bytes), nil
}
