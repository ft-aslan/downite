package tracker

import (
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

type AnnounceResponse struct {
	Interval uint64
	Peers    []PeerAddress
}
type PeerAddress struct {
	Ip   string
	Port uint16
}

type Tracker struct {
	Interval uint64
	Url      *url.URL
	Peers    []PeerAddress
}

func New(trackerUrl *url.URL) (*Tracker, error) {
	res, err := http.Get(trackerUrl.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	announceRes := AnnounceResponse{}
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
