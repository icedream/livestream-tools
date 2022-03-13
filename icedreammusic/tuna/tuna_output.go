package tuna

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TunaPlaybackStatus string

const (
	Stopped TunaPlaybackStatus = "stopped"
	Paused                     = "paused"
	Playing                    = "playing"
)

type TunaOutput struct {
	client *http.Client
}

type TunaData struct {
	// CoverURL is a URL to the track's cover art.
	CoverURL string `json:"cover_url,omitempty"`

	// Title is the track's title.
	Title string `json:"title,omitempty"`

	// Artists lists the artists of the track.
	Artists []string `json:"artists,omitempty"`

	// Album is the track's album name.
	Album string `json:"album,omitempty"`

	// Explicit determines whether this track is marked as containing explicit lyrics.
	Explicit bool `json:"explicit"`

	// DiscNumber is the track's disc number.
	DiscNumber int `json:"disc_number,omitempty"`

	// TrackNumber is the track's number on the disc.
	TrackNumber int `json:"track_number,omitempty"`

	// Year is the year of this track's release.
	Year int `json:"year,omitempty"`

	// Month is the month of this track's release.
	Month uint8 `json:"month,omitempty"`

	// Day is the day of this track's release.
	Day uint8 `json:"day,omitempty"`

	// Label is the publisher/label of the track.
	Label string `json:"label,omitempty"`

	// Status is the current state of track playback.
	Status TunaPlaybackStatus `json:"status"`

	// Progress is how much of the track has been played back in milliseconds.
	Progress uint64 `json:"progress"`

	// Duration is the duration of the track in milliseconds.
	Duration uint64 `json:"duration,omitempty"`

	// TimeLeft is how much of the track is left to play in milliseconds.
	TimeLeft float64 `json:"time_left,omitempty"`
}

func (d *TunaData) Equal(other *TunaData) bool {
	result := fmt.Sprintf("%+v", d) == fmt.Sprintf("%+v", other)
	return result
}

type TunaRequest struct {
	Data     TunaData `json:"data"`
	Hostname string   `json:"hostname,omitempty"`
	Date     uint64   `json:"date"`
}

func NewTunaOutput() *TunaOutput {
	return &TunaOutput{
		client: &http.Client{
			Timeout: time.Second * 2,
		},
	}
}

func (output *TunaOutput) Post(data *TunaData) (err error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(&TunaRequest{
		Data: *data,
		Date: uint64(time.Now().UnixMilli()),
	})
	_, err = output.client.Post("http://localhost:1608", "application/json", body)
	return
}
