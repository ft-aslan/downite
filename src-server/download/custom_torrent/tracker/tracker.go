package tracker

import (
	"downite/download/custom_torrent/decoding"
	"downite/download/custom_torrent/peer"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

type AnnounceResponse struct {
	Interval uint64 `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Tracker struct {
	Interval uint64
	Url      *url.URL
	Peers    []peer.PeerAddress
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

	peers, err := decoding.UnmarshalPeers([]byte(announceRes.Peers))
	if err != nil {
		return nil, err
	}

	tracker := Tracker{
		Interval: announceRes.Interval,
		Url:      trackerUrl,
		Peers:    peers,
	}
	return &tracker, nil
}
