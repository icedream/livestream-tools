package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"

	"github.com/icedream/livestream-tools/icedreammusic/metacollector"
	"github.com/icedream/livestream-tools/icedreammusic/tuna"
)

func main() {
	c, fs := NewNowPlayingFilesystem()
	host := fuse.NewFileSystemHost(fs)
	host.SetCapReaddirPlus(true)

	r := gin.Default()
	r.GET("/cover/:base64Path", func(c *gin.Context) {
		path := c.Params.ByName("base64Path")
		pathBytes, err := base64.URLEncoding.DecodeString(path)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
			return
		}
		path = string(pathBytes)

		f, err := os.Open(path)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
			return
		}
		defer f.Close()

		// get cover if possible
		fileMetadata, err := tag.ReadFrom(f)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
			return
		}

		picture := fileMetadata.Picture()
		if picture == nil {
			c.JSON(404, map[string]string{"error": "this file has no picture"})
			return
		}
		c.Header("Content-type", picture.MIMEType)
		c.Writer.Write(picture.Data)
		return
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	apiAddr := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", "127.0.0.1", listener.Addr().(*net.TCPAddr).Port),
		Path:   "/",
	}
	go http.Serve(listener, r)

	go func() {
		tunaOutput := tuna.NewTunaOutput()

		lastCoverCheckPath := ""
		lastCoverCheckResult := false
		lastCoverCheckTime := time.Now()

		metacollectorClient := metacollector.NewMetaCollectorClient(&url.URL{
			Scheme: "http",
			Host:   "192.168.188.69:8080", // TODO - make configurable
			Path:   "/",
		})

		for metadata := range c {
			// log.Printf("New metadata: %+v", metadata)

			status := "stopped"
			if metadata.IsPlaying {
				status = "playing"
			}

			tunaMetadata := &tuna.TunaData{
				Title:    metadata.Title,
				Artists:  []string{metadata.Artist},
				Label:    metadata.Publisher,
				Status:   status,
				Duration: uint64(metadata.Length * 1000),
				Progress: uint64(metadata.PlaybackTime * 1000),
			}

			if metadata.IsPlaying {
				hasChanged := lastCoverCheckPath != metadata.Path
				fi, err := os.Stat(metadata.Path)
				if err == nil {
					if !hasChanged {
						hasChanged = fi.ModTime().Sub(lastCoverCheckTime) > 0
					}
					lastCoverCheckTime = fi.ModTime()

					if hasChanged {
						lastCoverCheckResult = false
						lastCoverCheckPath = metadata.Path
						f, err := os.Open(metadata.Path)
						if err == nil {
							// get cover if possible
							fileMetadata, err := tag.ReadFrom(f)
							if err == nil {
								if fileMetadata.Picture() != nil {
									lastCoverCheckResult = true
								}
							} else {
								log.Printf("Warning while reading tags for %s: %s", metadata.Path, err)
							}
							f.Close()
						} else {
							log.Printf("Warning while opening file %s: %s", metadata.Path, err)
						}
					}
				} else {
					log.Printf("Warning while stat'ing file %s: %s", metadata.Path, err)
				}

				if lastCoverCheckResult {
					tunaMetadata.CoverURL = apiAddr.ResolveReference(&url.URL{
						Path: fmt.Sprintf("cover/%s", base64.URLEncoding.EncodeToString([]byte(metadata.Path))),
					}).String()
				}
			}

			go func() {
				// enrich metadata with metacollector
				resp, err := metacollectorClient.GetTrack(metacollector.MetaCollectorRequest{
					Artist: tunaMetadata.Artist,
					Title:  tunaMetadata.Title,
				})
				if err == nil {
					if resp.CoverURL != nil {
						tunaMetadata.CoverURL = *resp.CoverURL
					}
					tunaMetadata.Label = resp.Publisher
				}

				err = tunaOutput.Post(tunaMetadata)
				if err != nil {
					log.Println(err)
				} /*else {
					log.Printf("Tuna has received the metadata: %+v", tunaMetadata)
				}*/
			}()
		}
	}()

	host.Mount("", os.Args[1:])

}
