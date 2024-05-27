package types

import (
	"encoding/json"
	"net"
	"net/url"

	gotorrent "github.com/anacrolix/torrent"
	gotorrenttypes "github.com/anacrolix/torrent/types"
)

var PiecePriorityStringMap = map[string]gotorrenttypes.PiecePriority{
	"none":    gotorrenttypes.PiecePriorityNone,
	"maximum": gotorrenttypes.PiecePriorityNow,
	"high":    gotorrenttypes.PiecePriorityHigh,
	"normal":  gotorrenttypes.PiecePriorityNormal,
}

type TorrentStatus int

const (
	TorrentStatusPaused TorrentStatus = iota
	TorrentStatusDownloading
	TorrentStatusCompleted
	TorrentStatusSeeding
	TorrentStatusMetadata
)

var TorrentStatusStringMap = map[TorrentStatus]string{
	TorrentStatusPaused:      "paused",
	TorrentStatusDownloading: "downloading",
	TorrentStatusCompleted:   "completed",
	TorrentStatusSeeding:     "seeding",
	TorrentStatusMetadata:    "metadata",
}

func (s TorrentStatus) String() string {
	return TorrentStatusStringMap[s]
}
func (s TorrentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type PieceProgress struct {
	Index               int `json:"index"`
	DownloadedByteCount int `json:"downloadedByteCount"`
	Length              int `json:"length"`
}
type TorrentFileTreeNode struct {
	Length   int64                   `json:"length"`
	Name     string                  `json:"name"`
	Priority string                  `json:"priority" enum:"none,low,normal,high,maximum"`
	Path     string                  `json:"path"`
	Children *[]*TorrentFileTreeNode `json:"children"`
}
type TorrentFileFlatTree struct {
	Name     string `json:"name"`
	Priority string `json:"priority" enum:"none,low,normal,high,maximum"`
	Path     string `json:"path"`
}
type TorrentMeta struct {
	TotalSize int64                  `json:"totalSize"`
	Files     []*TorrentFileTreeNode `json:"files"`
	Name      string                 `json:"name"`
	Infohash  string                 `json:"infohash"`
	Magnet    string                 `json:"magnet"`
}
type Torrent struct {
	Name          string                        `json:"name"`
	Infohash      string                        `json:"infohash"`
	Files         []*TorrentFileTreeNode        `json:"files"`
	TotalSize     int64                         `json:"totalSize" db:"total_size"`
	AmountLeft    int64                         `json:"amountLeft"`
	Uploaded      int64                         `json:"uploaded"`
	Downloaded    int64                         `json:"downloaded"`
	Magnet        string                        `json:"magnet"`
	Status        string                        `json:"status" enum:"paused,downloading,completed,seeding,metadata"`
	PieceProgress []PieceProgress               `json:"pieceProgress"`
	Peers         map[string]gotorrent.PeerInfo `json:"peers"`
	Progress      float32                       `json:"progress"`
	PeerCount     int                           `json:"peerCount"`
	Eta           int                           `json:"eta"`
	CategoryId    int                           `json:"-" db:"category_id"`
	Category      string                        `json:"category"`
	SavePath      string                        `json:"savePath" db:"save_path"`
	Tags          []string                      `json:"tags"`
	Trackers      []Tracker                     `json:"trackers"`
	CreatedAt     int64                         `json:"createdAt" db:"created_at"`
	StartedAt     int64                         `json:"startedAt" db:"started_at"`
	TimeActive    int64                         `json:"timeActive" db:"time_active"`
	Availability  float32                       `json:"availability"`
	Ratio         float32                       `json:"ratio"`
	Seeds         int                           `json:"seeds"`
	DownloadSpeed float32                       `json:"downloadSpeed"`
	UploadSpeed   int                           `json:"uploadSpeed"`
	Comment       string                        `json:"comment"`
}
type Tracker struct {
	Interval uint64 `json:"interval"`
	Url      string `json:"url"`
	Peers    []Peer `json:"peers"`
	Tier     int    `json:"tier"`
}
type Peer struct {
	Url    url.URL `json:"url"`
	Ip     net.IP  `json:"ip"`
	IpPort uint16  `json:"ipPort"`
}
type TorrentClientConfig struct {
	PieceCompletionDbPath string
	DownloadPath          string
}
