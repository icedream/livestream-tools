package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type TunaOutput struct {
	client *http.Client
}

type tunaData struct {
	CoverURL string   `json:"cover_url"`
	Title    string   `json:"title"`
	Artists  []string `json:"artists"`
	Status   string   `json:"status"`
	Progress float64  `json:"progress"`
	Duration float64  `json:"duration"`
}

func NewTunaOutput() *TunaOutput {
	return &TunaOutput{
		client: http.DefaultClient,
	}
}

func (output *TunaOutput) Post(data *tunaData) (err error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(&struct {
		Data     *tunaData `json:"data"`
		Hostname string    `json:"hostname,omitempty"`
		Date     string    `json:"date"`
	}{
		Data: data,
		Date: time.Now().Format(time.RFC3339),
	})
	_, err = output.client.Post("http://localhost:1608", "application/json", body)
	return
}
