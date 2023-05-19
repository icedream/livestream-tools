package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/icedream/livestream-tools/icedreammusic/tuna"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
)

var (
	cli = kingpin.New("tunaposter", "Retrieve and copy Tuna now playing information to a Liquidsoap metadata Harbor endpoint.")

	argLiquidsoapMetaEndpointURL = cli.Arg("liquidsoap-meta-endpoint-url", "Liquidsoap metadata harbor endpoint URL").Required().URL()
	argTunaWebServerURL          = cli.Arg("tuna-webserver-url", "Tuna webserver URL").Default("http://localhost:1608").URL()
)

type liquidsoapMetadataRequest struct {
	Data liquidsoapMetadata `json:"data"`
}

type liquidsoapMetadata struct {
	CoverURL             string `json:"cover_url,omitempty"`
	MetadataBlockPicture string `json:"metadata_block_picture,omitempty"`
	Artist               string `json:"artist,omitempty"`
	Title                string `json:"title"`
	Publisher            string `json:"publisher,omitempty"`
	Year                 string `json:"year,omitempty"`
	Duration             uint64 `json:"duration,omitempty"`
	Progress             uint64 `json:"progress,omitempty"`
}

func (lm *liquidsoapMetadata) SetCover(r io.Reader, compressToJPEG bool) (err error) {
	description := ""

	// prepare data for reuse
	imageBuffer := new(bytes.Buffer)
	if _, err = io.Copy(imageBuffer, r); err != nil {
		return
	}
	imageBytes := imageBuffer.Bytes()

	// parse image metadata
	decodedImage, imageFormatName, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return
	}
	mime := ""
	switch imageFormatName {
	case "jpeg":
		mime = "image/jpeg"
	case "png":
		mime = "image/png"
	default:
		err = image.ErrFormat
	}

	// compress image if wanted
	if compressToJPEG && imageFormatName != "jpeg" {
		imageBuffer = new(bytes.Buffer)
		if err = jpeg.Encode(imageBuffer, decodedImage, &jpeg.Options{
			Quality: 75,
		}); err != nil {
			return
		}
		mime = "image/jpeg"
	}

	// Build METADATA_BLOCK_PICTURE
	// https://xiph.org/flac/format.html#metadata_block_picture

	w := new(strings.Builder)
	wb64 := base64.NewEncoder(base64.StdEncoding, w)
	binary.Write(wb64, binary.BigEndian, uint32(3))                          // type: cover (front)
	binary.Write(wb64, binary.BigEndian, uint32(len(mime)))                  // mime length
	wb64.Write([]byte(mime))                                                 // mime
	binary.Write(wb64, binary.BigEndian, uint32(len(description)))           // description length
	wb64.Write([]byte(description))                                          // description
	binary.Write(wb64, binary.BigEndian, uint32(decodedImage.Bounds().Dx())) // pixel width
	binary.Write(wb64, binary.BigEndian, uint32(decodedImage.Bounds().Dy())) // pixel height

	// color depth and paletted color count
	var bpp uint32
	var colorsUsed uint32
	switch v := decodedImage.(type) {
	case *image.Gray:
		bpp = 8
	case *image.Paletted:
		bpp = 8
		colorsUsed = uint32(len(v.Palette))
	case *image.RGBA:
		if v.Opaque() {
			bpp = 24
		} else {
			bpp = 32
		}
	case *image.NRGBA:
		if v.Opaque() {
			bpp = 24
		} else {
			bpp = 32
		}
	default:
		bpp = 24
	}
	binary.Write(wb64, binary.BigEndian, bpp)
	binary.Write(wb64, binary.BigEndian, colorsUsed)

	binary.Write(wb64, binary.BigEndian, uint32(len(imageBytes))) // raw image size
	wb64.Write(imageBytes)

	wb64.Close()
	lm.MetadataBlockPicture = w.String()
	return
}

func init() {
	kingpin.MustParse(cli.Parse(os.Args[1:]))
}

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
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
				differentSongReceived := oldTunaData == nil ||
					oldTunaData.Title != tunaData.Title ||
					len(oldTunaData.Artists) != len(tunaData.Artists)
				differentDataReceived := differentSongReceived ||
					oldTunaData.Progress != tunaData.Progress ||
					oldTunaData.Duration != tunaData.Duration
				if !differentDataReceived {
					for i, artist := range oldTunaData.Artists {
						differentDataReceived = differentDataReceived || artist != tunaData.Artists[i]
					}
				}
				if differentDataReceived && tunaData.Artists != nil && len(tunaData.Artists) > 0 && len(tunaData.Title) > 0 {
					liquidsoapMetadata := &liquidsoapMetadata{
						Artist:    strings.Join(tunaData.Artists, ", "),
						CoverURL:  tunaData.CoverURL,
						Publisher: tunaData.Label,
						Title:     tunaData.Title,
						Duration:  tunaData.Duration,
						Progress:  tunaData.Progress,
					}

					if tunaData.Year > 0 {
						liquidsoapMetadata.Year = fmt.Sprintf("%d", tunaData.Year)
					}

					// transfer cover to liquidsoap metadata
					if differentSongReceived {
						if coverURL, err := url.Parse(tunaData.CoverURL); err == nil {
							if strings.EqualFold(coverURL.Scheme, "http") ||
								strings.EqualFold(coverURL.Scheme, "https") {
								log.Println("Downloading cover:", tunaData.CoverURL)
								resp, err := http.Get(tunaData.CoverURL)
								if err == nil {
									err = liquidsoapMetadata.SetCover(resp.Body, true)
									resp.Body.Close()
									if err != nil {
										log.Println("WARNING: Failed to transfer cover to liquidsoap metadata, skipping:", err.Error())
									}
								}

								// remove reference to localhost/127.*.*.*
								localhost := coverURL.Host == "localhost" || strings.HasSuffix(coverURL.Host, ".localhost")
								if !localhost {
									if ip := net.ParseIP(coverURL.Host); ip != nil {
										localhost = ip[0] == 127
									}
								}
								if localhost {
									liquidsoapMetadata.CoverURL = ""
								}
							}
						}
					}

					liquidsoapData := &liquidsoapMetadataRequest{
						Data: *liquidsoapMetadata,
					}

					postBuf := new(bytes.Buffer)
					jsonEncoder := json.NewEncoder(postBuf)
					jsonEncoder.SetEscapeHTML(false)
					if err = jsonEncoder.Encode(liquidsoapData); err == nil {
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
