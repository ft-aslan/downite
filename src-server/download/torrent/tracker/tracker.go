package tracker

import (
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

type AnnounceResponse struct {
	Interval uint64        `bencode:"interval"`
	Peers    []PeerAddress `bencode:"peers"`
}
type PeerAddress struct {
	Ip   string `bencode:"ip"`
	Port uint16 `bencode:"port"`
}

type Tracker struct {
	Interval uint64
	Url      *url.URL
	Peers    []PeerAddress
}

func New(trackerUrl *url.URL) (*Tracker, error) {
	trackerUrlString := trackerUrl.String()
	res, err := http.Get(trackerUrlString)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	announceRes := &AnnounceResponse{}
	err = bencode.Unmarshal(res.Body, announceRes)
	if err != nil {
		return nil, err
	}

	tracker := Tracker{
		Interval: announceRes.Interval,
		Url:      trackerUrl,
		Peers:    announceRes.Peers,
	}
	return &tracker, nil
}
