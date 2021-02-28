package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/icedream/livestream-tools/icedreammusic/tuna"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
)

var (
	cli = kingpin.New("tunaposter", "Retrieve and copy Tuna now playing information to a Liquidsoap metadata Harbor endpoint.")

	argTunaWebServerURL          = cli.Arg("tuna-webserver-url", "Tuna webserver URL").Required().URL()
	argLiquidsoapMetaEndpointURL = cli.Arg("liquidsoap-meta-endpoint-url", "Liquidsoap metadata harbor endpoint URL").Required().URL()
)

type liquidsoapMetadataRequest struct {
	Data liquidsoapMetadata `json:"data"`
}

type liquidsoapMetadata struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func init() {
	kingpin.MustParse(cli.Parse(os.Args[1:]))
}

func main() {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// TODO - shutdown signal handling
	var oldTunaData *tuna.TunaData

	for {
		// client.Get(url string)
		resp, err := client.Get((*argTunaWebServerURL).String())
		if err == nil {
			tunaData := new(tuna.TunaData)
			if err = json.NewDecoder(resp.Body).Decode(tunaData); err == nil {
				// skip empty or same metadata
				differentDataReceived := oldTunaData == nil ||
					oldTunaData.Title != tunaData.Title ||
					len(oldTunaData.Artists) != len(tunaData.Artists)
				if !differentDataReceived {
					for i, artist := range oldTunaData.Artists {
						differentDataReceived = differentDataReceived || artist != tunaData.Artists[i]
					}
				}
				if differentDataReceived && tunaData.Artists != nil && len(tunaData.Artists) > 0 && len(tunaData.Title) > 0 {
					liquidsoapData := &liquidsoapMetadataRequest{
						Data: liquidsoapMetadata{
							Artist: strings.Join(tunaData.Artists, ", "),
							Title:  tunaData.Title,
						},
					}
					postBuf := new(bytes.Buffer)
					if err = json.NewEncoder(postBuf).Encode(liquidsoapData); err == nil {
						postBufCopy := postBuf.Bytes()
						log.Println("Will send new metadata:", string(postBufCopy))
						if _, err = client.Post((*argLiquidsoapMetaEndpointURL).String(), "application/json", bytes.NewReader(postBufCopy)); err == nil {
							oldTunaData = tunaData
						} else {
							log.Printf("WARNING: Failed to post metadata to Liquidsoap harbor endpoint: %s", err.Error())
						}
					} else {
						log.Printf("WARNING: Failed to encode metadata for Liquidsoap harbor endpoint: %s", err.Error())
					}
				}
			} else {
				log.Printf("WARNING: Failed to decode metadata from Tuna webserver: %s", err.Error())
			}
		} else {
			log.Printf("WARNING: Failed to retrieve metadata from Tuna webserver, resetting old data: %s", err.Error())
			oldTunaData = nil
		}
		time.Sleep(time.Second)
	}
}
