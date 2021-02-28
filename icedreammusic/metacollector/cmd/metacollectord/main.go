package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/dhowden/tag"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/icedream/livestream-tools/icedreammusic/metacollector"
)

const (
	appID        = "metacollector"
	appName      = "Metadata collector"
	appEnvPrefix = appID

	configDatabase     = "Database"
	configDatabaseType = configDatabase + ".Type"
	configDatabaseURL  = configDatabase + ".URL"

	configLibrary      = "Library"
	configLibraryPaths = configLibrary + ".Paths"

	configServer        = "Server"
	configServerAddress = configServer + ".Address"
)

type track struct {
	gorm.Model
	Artist      string `gorm:"index:idx_artist_title"`
	Title       string `gorm:"index:idx_artist_title"`
	Publisher   string
	CoverFile   *file `gorm:"ForeignKey:CoverFileID;"`
	CoverFileID *uint
}

type file struct {
	gorm.Model
	Data        []byte
	ContentType string
}

type manager struct {
	database *gorm.DB
}

func newManager(appDatabase *gorm.DB) *manager {
	m := &manager{
		database: appDatabase.Debug(),
	}

	return m
}

func (m *manager) Migrate() error {
	if err := m.database.AutoMigrate(&file{}); err != nil {
		return err
	}
	if err := m.database.AutoMigrate(&track{}); err != nil {
		return err
	}
	return nil
}

func (m *manager) GetFile(id uint) (result *file, err error) {
	result = new(file)
	err = m.database.Find(result, id).Error
	return
}

func (m *manager) UploadFile(file *file) (err error) {
	err = m.database.Save(file).Error
	return
}

func (m *manager) DeleteFile(file *file) (err error) {
	err = m.database.Delete(file).Error
	return
}

func getPublisherFromTags(tags tag.Metadata) string {
	for _, tagName := range []string{
		// ID3v2
		"PUBLISHER",
		"GROUPING",
		"TPUB",
		// VORBIS
		"organization",
		"publisher",
		"grouping",
	} {
		if value, ok := tags.Raw()[tagName]; ok {
			if valueString, ok := value.(string); ok && len(strings.TrimSpace(valueString)) > 0 {
				return valueString
			}
		}
	}
	return ""
}

func (m *manager) UpdateFileFromFilesystem(f *os.File) (err error) {
	tags, err := tag.ReadFrom(f)
	if err != nil {
		return
	}

	// Try to find old information we can reuse
	trackObj, err := m.GetTrackByArtistAndTitle(tags.Artist(), tags.Title(), false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		trackObj = new(track)
	} else if err != nil {
		return // something went terribly wrong
	}

	trackObj.Artist = tags.Artist()
	trackObj.Title = tags.Title()

	// Cover
	if tags.Picture() != nil {
		trackObj.CoverFile = new(file)
		if trackObj.CoverFileID != nil {
			trackObj.CoverFile.ID = *trackObj.CoverFileID
		}
		trackObj.CoverFile.ContentType = tags.Picture().MIMEType
		trackObj.CoverFile.Data = tags.Picture().Data
	}

	// Publisher
	publisher := getPublisherFromTags(tags)
	if len(publisher) > 0 {
		trackObj.Publisher = publisher
	}

	m.database.Save(trackObj)

	return
}

func (m *manager) WriteTrack(track *track) (err error) {
	// If we got no ID, try to find out first whether we already have an entry with same title and artist
	result, err := m.GetTrackByArtistAndTitle(track.Artist, track.Title, false)
	if err == nil {
		// Found one!
		track.ID = result.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Something went really wrong here…
		return
	} else {
		err = nil
	}

	err = m.database.Save(track).Error
	return
}

func (m *manager) GetTrackByArtistAndTitle(artist, title string, withCoverFile bool) (result *track, err error) {
	result = new(track)
	db := m.database
	if withCoverFile {
		db = db.Preload("CoverFile")
	}
	err = db.
		Where("artist = @artist AND title COLLATE UTF8_GENERAL_CI LIKE @title",
			sql.Named("artist", artist),
			sql.Named("title", title+"%")).
		Find(result).Error
	return
}

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Panic(err)
		}
	}()

	viper.AddConfigPath("/etc/" + appID + "/")
	viper.AddConfigPath("$HOME/.config/" + appID + "/")
	viper.AddConfigPath("$HOME/." + appID + "/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.SetEnvPrefix(appEnvPrefix)
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			err = nil
			log.Println("No configuration file found, ignoring")
		} else {
			// Config file was found but another error was produced
			log.Panicf("Failed to read configuration: %s", err.Error())
		}
	}

	viper.SetDefault(configDatabaseType, "sqlite")
	viper.SetDefault(configDatabaseURL, "app.db")
	viper.SetDefault(configLibraryPaths, []string{})

	viper.Debug()

	var dialector gorm.Dialector
	switch viper.GetString(configDatabaseType) {
	case "sqlite":
		dialector = sqlite.Open(viper.GetString(configDatabaseURL))
	case "mysql":
		dialector = mysql.New(
			mysql.Config{
				DSN: viper.GetString(configDatabaseURL),
			})
	default:
		err = fmt.Errorf("Unsupported config database type: %s", viper.GetString(configDatabaseType))
		return
	}
	gormDatabase, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return
	}

	m := newManager(gormDatabase)
	if err = m.Migrate(); err != nil {
		return
	}

	// Handle shutdown gracefully
	ctx := context.Background() // TODO - Timeouts and fancy stuff
	signalChannel := make(chan os.Signal, 1)
	cond := sync.NewCond(&sync.Mutex{})
	go signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChannel
		cond.L.Lock()
		cond.Broadcast()
		cond.L.Unlock()
	}()

	wg := new(sync.WaitGroup)

	// Set up HTTP server
	r := gin.Default()
	r.GET("/file/:fileID", func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("fileID"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, "invalid file ID")
			return
		}
		file, err := m.GetFile(uint(fileID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.Header("Content-type", file.ContentType)
		c.Writer.Write(file.Data)
	})
	r.POST("/track/find", func(c *gin.Context) {
		var form struct {
			Artist string
			Title  string
		}
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, "missing POST data")
			return
		}
		if len(form.Artist) <= 0 {
			c.JSON(http.StatusBadRequest, "artist must not be empty")
			return
		}
		if len(form.Title) <= 0 {
			c.JSON(http.StatusBadRequest, "title must not be empty")
			return
		}
		track, err := m.GetTrackByArtistAndTitle(form.Artist, form.Title, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		response := metacollector.MetaCollectorResponse{
			Artist:    track.Artist,
			Title:     track.Title,
			Publisher: track.Publisher,
		}
		if track.CoverFileID != nil {
			coverURL := fmt.Sprintf("/file/%d", *track.CoverFileID)
			response.CoverURL = &coverURL
		}
		c.JSON(http.StatusOK, response)
	})
	server := new(http.Server)
	server.Addr = viper.GetString(configServerAddress)
	server.Handler = r
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
		if err := server.Shutdown(ctx); err != nil {
			log.Print("During server shutdown:", err)
		}
	}()

	// Watch library paths
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	wg.Add(1)
	go func() {
		defer wg.Done()
		quitSignalChannel := make(chan interface{}, 1)
		go func() {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
			quitSignalChannel <- nil
		}()
		for {
			select {
			case <-quitSignalChannel:
				watcher.Close()
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					f, err := os.Open(event.Name)
					if err == nil {
						err = m.UpdateFileFromFilesystem(f)
					}
					if err != nil {
						log.Printf("Failed to update tags from file system for %s: %s", event.Name, err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	firstScanQuitSignalChannel := make(chan interface{}, 1)
	go func() {
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
		firstScanQuitSignalChannel <- nil
	}()
	for _, watchPath := range viper.GetStringSlice(configLibraryPaths) {
		if len(strings.TrimSpace(watchPath)) <= 0 {
			continue // ignore empty path entries
		}

		log.Println("Adding path to watcher:", watchPath)
		err = watcher.Add(watchPath)
		if err != nil {
			return
		}

		// Force first scan
		filepath.WalkDir(watchPath, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			select {
			case <-firstScanQuitSignalChannel:
				return nil // just quit out asap
			default:
			}

			// skip directory entries
			if d.IsDir() {
				return nil
			}

			switch strings.ToLower(filepath.Ext(filePath)) {
			case ".ogg", ".mp3", ".m4a", ".aac", ".flac", ".wav", ".wma", ".wv":
				log.Println("scanning:", filePath)
				f, err := os.Open(filePath)
				if err == nil {
					err = m.UpdateFileFromFilesystem(f)
				}
				if err != nil {
					log.Printf("Failed to update tags from file system for %s: %s", filePath, err)
				}
			}
			return nil
		})
	}

	wg.Wait()
}
