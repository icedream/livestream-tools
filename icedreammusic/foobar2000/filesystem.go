package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	fuse "github.com/billziss-gh/cgofuse/fuse"
)

type NowPlayingMetadata struct {
	IsPlaying             bool
	PlaybackTime          float64
	PlaybackTimeRemaining float64
	Length                float64
	Path                  string
	Samplerate            int
	LengthSamples         int64
	Title                 string
	Artist                string
	Album                 string
	Publisher             string
}

type NowPlayingFilesystem struct {
	*Memfs

	lastMetadata NowPlayingMetadata
	metadataC    chan NowPlayingMetadata
}

func NewNowPlayingFilesystem() (c <-chan NowPlayingMetadata, i fuse.FileSystemInterface) {
	ch := make(chan NowPlayingMetadata)
	c = ch
	npfs := &NowPlayingFilesystem{
		Memfs:     NewMemfs(),
		metadataC: ch,
	}
	npfs.Memfs.Mkdir("nowplaying", 0777)
	// err, fh := npfs.Open("/nowplaying.json", 0)
	// if err != 0 {
	// 	panic(fmt.Errorf("Failed to create nowplaying.json in memory, error code %d", err))
	// }
	// npfs.Write("/nowplaying.json", []byte("{}"), 0, fh)
	// npfs.Release("/nowplaying.json", fh)

	i = npfs
	return
}

func (self *NowPlayingFilesystem) Release(path string, fh uint64) (retval int) {
	retval = self.Memfs.Release(path, fh)
	if retval != 0 {
		return
	}

	if filepath.Base(path) != "nowplaying.json" {
		return
	}

	errC, fh := self.Memfs.Open(path, 0)
	if errC != 0 {
		retval = errC
		return
	}
	defer self.Memfs.Release(path, fh)
	buff := make([]byte, 1024000)
	self.Memfs.Read(path, buff, 0, fh)
	metadata := new(NowPlayingMetadata)
	if err := json.NewDecoder(bytes.NewReader(buff)).Decode(metadata); err == nil {
		self.metadataC <- *metadata
	}
	return
}
