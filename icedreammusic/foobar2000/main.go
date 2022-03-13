package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v3-unstable"

	"github.com/icedream/livestream-tools/icedreammusic/metacollector"
	"github.com/icedream/livestream-tools/icedreammusic/tuna"
)

var (
	cli = kingpin.New("foobar2000", "Transmit foobar2000 now playing data to Tuna.")

	argMetacollectorURL = cli.Arg("metacollector-url", "Metadata collector URL (service normally runs on port 8080)").Required().URL()
	argDrive            = cli.Arg("mountpoint", "The mountpoint to attach to.").Default("Z:").String()
)

func init() {
	kingpin.MustParse(cli.Parse(os.Args[1:]))
}

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

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	go func() {
		tunaOutput := tuna.NewTunaOutput()

		lastCoverCheckPath := ""
		lastCoverCheckResult := false
		lastCoverCheckTime := time.Now()

		metacollectorClient := metacollector.NewMetaCollectorClient(*argMetacollectorURL)

		for metadata := range c {
			// log.Printf("New metadata: %+v", metadata)

			status := tuna.Stopped
			if metadata.IsPlaying {
				status = tuna.Playing
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
				// is this playing from Spotify?
				if strings.HasPrefix(metadata.Path, "spotify:") {
					// Spotify covers have to be fetched via their API
					// TODO - use token for API instead of old oembed way, see https://stackoverflow.com/a/18294883
					u := &url.URL{
						Scheme: "https",
						Host:   "open.spotify.com",
						Path:   "/oembed",
					}
					q := u.Query()
					q.Add("url", metadata.Path)
					u.RawQuery = q.Encode()
					if resp, err := httpClient.Get(u.String()); err == nil {
						oembedData := &struct {
							ThumbnailURL string `json:"thumbnail_url"`
						}{}
						if err = json.NewDecoder(resp.Body).Decode(oembedData); err == nil {
							tunaMetadata.CoverURL = oembedData.ThumbnailURL
						}
					}
				} else {
					// Check normal/other files against metacollector
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
			}

			go func() {
				// enrich metadata with metacollector
				req := metacollector.MetaCollectorRequest{
					Artist: tunaMetadata.Artists[0],
					Title:  tunaMetadata.Title,
				}
				log.Printf("Trying to enrich metadata: %+v", req)
				resp, err := metacollectorClient.GetTrack(req)
				if err == nil {
					log.Println("Enriching metadata:", resp)
					if resp.CoverURL != nil {
						tunaMetadata.CoverURL = (*argMetacollectorURL).ResolveReference(&url.URL{
							Path: *resp.CoverURL,
						}).String()
					}
					tunaMetadata.Label = resp.Publisher
				} else {
					log.Println("Failed to enrich metadata:", err)
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

	host.Mount("", []string{
		*argDrive,
	})
}
