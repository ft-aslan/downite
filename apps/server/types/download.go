package types

import (
	"database/sql"
	"time"
)

type DownloadStatus int

const (
	DownloadStatusPaused DownloadStatus = iota
	DownloadStatusDownloading
	DownloadStatusCompleted
	DownloadStatusError
	DownloadStatusMetadata
)

var DownloadStatusStringMap = map[DownloadStatus]string{
	DownloadStatusPaused:      "paused",
	DownloadStatusDownloading: "downloading",
	DownloadStatusCompleted:   "completed",
	DownloadStatusError:       "error",
	DownloadStatusMetadata:    "metadata",
}

func (d DownloadStatus) String() string {
	return DownloadStatusStringMap[d]
}

type Download struct {
	Id                  int             `json:"id"`
	CreatedAt           time.Time       `json:"createdAt" db:"created_at"`
	StartedAt           sql.NullTime    `json:"startedAt" db:"started_at"`
	TimeActive          time.Duration   `json:"timeActive" db:"time_active"`
	FinishedAt          sql.NullTime    `json:"finishedAt" db:"finished_at"`
	Status              string          `json:"status" enum:"paused,downloading,completed,error,metadata"`
	Name                string          `json:"name"`
	SavePath            string          `db:"save_path" json:"savePath"`
	PartCount           int             `db:"part_count" json:"partCount"`
	PartLength          uint64          `db:"part_length" json:"partLength"`
	TotalSize           uint64          `db:"total_size" json:"totalSize"`
	DownloadedBytes     uint64          `db:"downloaded_bytes" json:"downloadedBytes"`
	BytesWritten        uint64          `db:"-" json:"-"`
	DownloadSpeed       uint64          `db:"-" json:"downloadSpeed"`
	Progress            float64         `json:"progress" db:"-"`
	Parts               []*DownloadPart `json:"parts" db:"-"`
	IsMultiPart         bool            `json:"isMultiPart" db:"is_multi_part"`
	Url                 string          `json:"url"`
	QueueNumber         int             `db:"queue_number" json:"queueNumber"`
	CurrentWrittenBytes uint64          `db:"-" json:"-"`
	Error               string          `db:"error" json:"error"`
}

func (download *Download) Write(bytes []byte) (int, error) {
	download.DownloadedBytes += uint64(len(bytes))
	download.BytesWritten += uint64(len(bytes))
	download.Progress = float64(download.DownloadedBytes) / float64(download.TotalSize) * 100
	// fmt.Printf("downloaded bytes : %d \n", download.DownloadedBytes)
	return len(bytes), nil
}

type DownloadPart struct {
	Id              int           `db:"id" json:"-"`
	CreatedAt       time.Time     `db:"created_at" json:"createdAt"`
	StartedAt       sql.NullTime  `db:"started_at" json:"startedAt"`
	TimeActive      time.Duration `db:"time_active" json:"timeActive"`
	FinishedAt      sql.NullTime  `db:"finished_at" json:"finishedAt"`
	Status          string        `json:"status"`
	PartIndex       int           `db:"part_index" json:"partIndex"`
	StartByteIndex  uint64        `db:"start_byte_index" json:"startByteIndex"`
	EndByteIndex    uint64        `db:"end_byte_index" json:"endByteIndex"`
	PartLength      uint64        `db:"part_length" json:"partLength"`
	DownloadedBytes uint64        `db:"downloaded_bytes" json:"downloadedBytes"`
	Progress        float64       `db:"-" json:"progress"`
	DownloadId      int           `json:"-" db:"download_id"`
	Error           string        `db:"error" json:"error"`
}

func (part *DownloadPart) Write(bytes []byte) (int, error) {
	part.DownloadedBytes += uint64(len(bytes))
	part.Progress = float64(part.DownloadedBytes) / float64(part.PartLength) * 100
	// fmt.Printf("downloaded bytes for part number %d : | bytes : %d \n", part.PartIndex, part.DownloadedBytes)
	return len(bytes), nil
}

type DownloadMeta struct {
	TotalSize          uint64 `json:"totalSize"`
	Url                string `json:"url"`
	FileName           string `json:"fileName"`
	FileType           string `json:"fileType"`
	IsRangeAllowed     bool   `json:"isRangeAllowed"`
	IsExist            bool   `json:"isExist"`
	ExistingDownloadId int    `json:"existingDownloadId"`
}
