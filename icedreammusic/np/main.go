package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/delthas/go-libnp"
	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v3-unstable"

	"github.com/icedream/livestream-tools/icedreammusic/metacollector"
	"github.com/icedream/livestream-tools/icedreammusic/tuna"
)

var (
	cli = kingpin.New("np", "Transmit system now playing data to Tuna.")

	argMetacollectorURL = cli.Arg("metacollector-url", "Metadata collector URL (service normally runs on port 8080)").Required().URL()
	argDrive            = cli.Arg("mountpoint", "The mountpoint to attach to.").Default("Z:").String()
)

func init() {
	kingpin.MustParse(cli.Parse(os.Args[1:]))
}

func watchMetadata(ctx context.Context) <-chan *libnp.Info {
	ticker := time.NewTicker(time.Second)
	c := make(chan *libnp.Info)
	go func(ticker *time.Ticker) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info, err := libnp.GetInfo(context.Background())
				if err != nil {
					os.Stderr.WriteString("WARNING: " + err.Error() + "\n")
					continue
				}
				c <- info
			}
		}
	}(ticker)

	return c
}

func generateIDFromMetadata(metadata libnp.Info) [64]byte {
	return sha512.Sum512([]byte(strings.Join(metadata.Artists, "|") + "||" + metadata.Title))
}

func main() {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := watchMetadata(ctx)

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

	func() {
		tunaOutput := tuna.NewTunaOutput()

		var lastCoverCheckPath [64]byte
		lastCoverCheckResult := false
		lastCoverCheckTime := time.Now()

		metacollectorClient := metacollector.NewMetaCollectorClient(*argMetacollectorURL)

		for metadata := range c {
			log.Printf("New metadata: %+v", metadata)

			tunaMetadata := &tuna.TunaData{
				Status: tuna.Stopped,
			}

			if metadata != nil &&
				(metadata.PlaybackType == libnp.PlaybackTypeMusic ||
					metadata.PlaybackType != libnp.PlaybackTypeVideo) {
				tunaMetadata.Status = tuna.Playing

				id := generateIDFromMetadata(*metadata)
				tunaMetadata.Title = metadata.Title
				tunaMetadata.Artists = metadata.Artists
				tunaMetadata.Duration = uint64(metadata.Length.Milliseconds())

				// Check normal/other files against metacollector
				hasChanged := lastCoverCheckPath != id
				trackURL, err := url.Parse(metadata.URL)
				if err == nil {
					if strings.EqualFold(trackURL.Scheme, "file") {
						fi, err := os.Stat(trackURL.Path)
						if err == nil {
							if !hasChanged {
								hasChanged = fi.ModTime().Sub(lastCoverCheckTime) > 0
							}
							lastCoverCheckTime = fi.ModTime()

							if hasChanged {
								lastCoverCheckResult = false
								lastCoverCheckPath = id
								f, err := os.Open(trackURL.Path)
								if err == nil {
									// get cover if possible
									fileMetadata, err := tag.ReadFrom(f)
									if err == nil {
										if fileMetadata.Picture() != nil {
											lastCoverCheckResult = true
										}
									} else {
										log.Printf("Warning while reading tags for %s: %s", trackURL.Path, err)
									}
									f.Close()
								} else {
									log.Printf("Warning while opening file %s: %s", trackURL.Path, err)
								}
							}
						} else {
							log.Printf("Warning while stat'ing file %s: %s", trackURL.Path, err)
						}

						if lastCoverCheckResult {
							tunaMetadata.CoverURL = apiAddr.ResolveReference(&url.URL{
								Path: fmt.Sprintf("cover/%s", base64.URLEncoding.EncodeToString([]byte(trackURL.Path))),
							}).String()
						}
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
}
