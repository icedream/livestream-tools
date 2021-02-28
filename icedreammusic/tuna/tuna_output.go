package tuna

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type TunaOutput struct {
	client *http.Client
}

type TunaData struct {
	CoverURL string   `json:"cover_url"`
	Title    string   `json:"title"`
	Artists  []string `json:"artists"`
	Label    string   `json:"label"`
	Status   string   `json:"status"`
	Progress uint64   `json:"progress"`
	Duration uint64   `json:"duration"`
}

func (d *TunaData) Equal(other *TunaData) bool {
	result := fmt.Sprintf("%+v", d) == fmt.Sprintf("%+v", other)
	log.Printf("%+v == %+v => %v", d, other, result)
	return result
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
	json.NewEncoder(body).Encode(&struct {
		Data     *TunaData `json:"data"`
		Hostname string    `json:"hostname,omitempty"`
		Date     string    `json:"date"`
	}{
		Data: data,
		Date: time.Now().Format(time.RFC3339),
	})
	_, err = output.client.Post("http://localhost:1608", "application/json", body)
	return
}
