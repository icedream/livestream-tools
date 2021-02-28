package main

import (
	"log"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/icedream/go-stagelinq"
	"github.com/icedream/livestream-tools/icedreammusic/metacollector"
	"github.com/icedream/livestream-tools/icedreammusic/tuna"
)

type ReceivedMetadata struct {
	Device *stagelinq.Device
	State  *stagelinq.State
}

type MultiMetadataTracker struct {
	lock             sync.Mutex
	token            stagelinq.Token
	metadataChannel  chan *ReceivedMetadata
	connectedDevices map[*stagelinq.Device]DeviceConnections
	waitGroup        sync.WaitGroup
}

type DeviceConnections struct {
	MainConn     *stagelinq.MainConnection
	StateMapConn net.Conn
}

func newMultiMetadataTracker(token stagelinq.Token) *MultiMetadataTracker {
	c := make(chan *ReceivedMetadata)
	return &MultiMetadataTracker{
		token:            token,
		metadataChannel:  c,
		connectedDevices: map[*stagelinq.Device]DeviceConnections{},
	}
}

func (mmt *MultiMetadataTracker) Stop(dev *stagelinq.Device) {
	defer mmt.synchronize()()
	for registeredDevice, conns := range mmt.connectedDevices {
		if registeredDevice.IsEqual(dev) {
			conns.StateMapConn.Close()
			conns.MainConn.Close()
			return
		}
	}
}

func (mmt *MultiMetadataTracker) Close() {
	for _, conn := range mmt.connectedDevices {
		conn.StateMapConn.Close()
		conn.MainConn.Close()
	}
	mmt.waitGroup.Wait()
	close(mmt.metadataChannel)
}

func (mmt *MultiMetadataTracker) synchronize() func() {
	mmt.lock.Lock()
	return func() {
		mmt.lock.Unlock()
	}
}

func (mmt *MultiMetadataTracker) Start(dev *stagelinq.Device) {
	mmt.registerDevice(dev)
}

func (mmt *MultiMetadataTracker) registerDevice(dev *stagelinq.Device) {
	defer mmt.synchronize()()

	// check if device was already added
	for registeredDevice := range mmt.connectedDevices {
		if registeredDevice.IsEqual(dev) {
			return
		}
	}

	log.Printf("Found %s %s (%s)", dev.SoftwareName, dev.SoftwareVersion, dev.Name)

	// try and connect to device
	devConn, err := dev.Connect(mmt.token, []*stagelinq.Service{})
	if err != nil {
		log.Printf("WARNING: Could not connect to %s: %s", dev.IP, err.Error())
		return
	}
	services, err := devConn.RequestServices()
	if err != nil {
		log.Printf("WARNING: Failed to retrieve services of %s: %s", dev.IP, err.Error())
		devConn.Close()
		return
	}
	for _, service := range services {
		if service.Port == 0 {
			continue
		}
		switch service.Name {
		case "StateMap":
			log.Printf("Connecting to %s:%d for %s...", dev.IP, service.Port, service.Name)
			rawConn, err := dev.Dial(service.Port)
			if err != nil {
				log.Printf("WARNING: Failed to connect to state map service at %s: %s", dev.IP, err.Error())
				return
			}
			log.Printf("Handshaking with %s:%d for %s...", dev.IP, service.Port, service.Name)
			stateMapConn, err := stagelinq.NewStateMapConnection(rawConn, mmt.token)
			if err != nil {
				log.Printf("WARNING: Failed to handshake state map connection at %s: %s", dev.IP, err.Error())
				rawConn.Close()
				devConn.Close()
				return
			}
			mmt.connectedDevices[dev] = DeviceConnections{
				MainConn:     devConn,
				StateMapConn: rawConn,
			}
			for _, key := range []string{
				stagelinq.EngineDeck1Play,
				stagelinq.EngineDeck1TrackArtistName,
				stagelinq.EngineDeck1TrackSongName,
				stagelinq.EngineDeck2Play,
				stagelinq.EngineDeck2TrackArtistName,
				stagelinq.EngineDeck2TrackSongName,
				stagelinq.EngineDeck3Play,
				stagelinq.EngineDeck3TrackArtistName,
				stagelinq.EngineDeck3TrackSongName,
				stagelinq.EngineDeck4Play,
				stagelinq.EngineDeck4TrackArtistName,
				stagelinq.EngineDeck4TrackSongName,
				stagelinq.MixerCH1faderPosition,
				stagelinq.MixerCH2faderPosition,
				stagelinq.MixerCH3faderPosition,
				stagelinq.MixerCH4faderPosition,
			} {
				stateMapConn.Subscribe(key)
			}
			mmt.trackStateMap(dev, stateMapConn)
		}
	}
}

func (mmt *MultiMetadataTracker) unregisterDevice(dev *stagelinq.Device) {
	log.Printf("About to unregister %s...", dev.IP)
	defer mmt.synchronize()()
	for registeredDevice, conns := range mmt.connectedDevices {
		if registeredDevice.IsEqual(dev) {
			conns.StateMapConn.Close()
			conns.MainConn.Close()
			delete(mmt.connectedDevices, registeredDevice)
			return
		}
	}
}

func (mmt *MultiMetadataTracker) trackStateMap(dev *stagelinq.Device, conn *stagelinq.StateMapConnection) {
	log.Printf("Tracking %s...", dev.IP)
	mmt.waitGroup.Add(1)
	go func() {
		defer mmt.waitGroup.Done()
		defer mmt.unregisterDevice(dev)
		for {
			select {
			case err := <-conn.ErrorC():
				log.Printf("WARNING: Disconnected from state map at %s: %s", dev.IP, err.Error())
				return
			case state := <-conn.StateC():
				mmt.metadataChannel <- &ReceivedMetadata{
					Device: dev,
					State:  state,
				}
			}
		}
	}()
}

func (mmt *MultiMetadataTracker) C() <-chan *ReceivedMetadata {
	return mmt.metadataChannel
}

func main() {
	listener, err := stagelinq.ListenWithConfiguration(&stagelinq.ListenerConfiguration{
		Name:            "icedreamnowplaying",
		SoftwareName:    "Icedream's Now Playing",
		SoftwareVersion: "0.0.0",
	})
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	listener.AnnounceEvery(time.Second)

	tracker := newMultiMetadataTracker(listener.Token())
	defer tracker.Close()

	// Device tracking
	go func() {
		for {
			device, deviceState, err := listener.Discover(0)
			if device.SoftwareName == "Icedream's Now Playing" {
				continue // found our own software
			}
			if err != nil {
				log.Printf("WARNING: During discovery an error occured: %s", err.Error())
				continue
			}
			switch deviceState {
			case stagelinq.DeviceLeaving:
				tracker.Stop(device)
			case stagelinq.DevicePresent:
				tracker.Start(device)
			}
		}
	}()

	// Actual metadata collection and analysis
	var currentMetadata *DeviceMeta
	var lastDetectedMeta *DeviceMeta
	metadata := map[*stagelinq.Device]map[int]*DeviceMeta{}

	output := tuna.NewTunaOutput()
	metaCollectorAPIURL := &url.URL{
		Scheme: "http",
		Host:   "192.168.188.69:8080", // TODO - make configurable
		Path:   "/",
	}
	metacollectorClient := metacollector.NewMetaCollectorClient(metaCollectorAPIURL)

	sendMetadata := func() {
		tunaData := &tuna.TunaData{
			Status: "stopped",
		}
		if currentMetadata != nil {
			tunaData.Status = "playing"
			tunaData.Artists = []string{currentMetadata.Artist}
			tunaData.Title = currentMetadata.Title
		}
		// enrich metadata with metacollector
		resp, err := metacollectorClient.GetTrack(metacollector.MetaCollectorRequest{
			Artist: currentMetadata.Artist,
			Title:  currentMetadata.Title,
		})
		if err == nil {
			if resp.CoverURL != nil {
				tunaData.CoverURL = metaCollectorAPIURL.ResolveReference(&url.URL{
					Path: *resp.CoverURL,
				}).String()
			}
			tunaData.Label = resp.Publisher
		}
		if err := output.Post(tunaData); err != nil {
			log.Printf("WARNING: Failed to send new metadata to tuna: %s", err.Error())
		}
	}
	conflictDetect := 0
	sameMeta := 0
	detectNewMetadata := func() {
		tracksRunning := 0
		var newFaderValue float64
		var maxVolumeDifference float64 = 2
		var newMeta *DeviceMeta
		for _, meta := range metadata {
			for _, deck := range meta {
				if !deck.Playing {
					continue
				}
				tracksRunning++
				if newMeta != nil {
					volumeDifference := newFaderValue - deck.Fader
					if volumeDifference < 0 {
						volumeDifference *= -1
					}
					if volumeDifference < maxVolumeDifference {
						maxVolumeDifference = volumeDifference
					}
					if deck.Fader < newFaderValue {
						continue
					}
					newFaderValue = deck.Fader
					newMeta = deck
				} else {
					newFaderValue = deck.Fader
					newMeta = deck
				}
			}
		}
		lastDetectedMeta = newMeta
		if maxVolumeDifference < 0.4 && tracksRunning > 1 {
			conflictDetect++
			if conflictDetect > 15 {
				currentMetadata = nil
			}
			sameMeta = 0
		} else {
			isSameMeta := (newMeta == nil && lastDetectedMeta == nil) || ((newMeta != nil && lastDetectedMeta != nil) && *newMeta == *lastDetectedMeta)
			if isSameMeta {
				sameMeta++
			} else {
				sameMeta = 0
			}
			if sameMeta > 10 {
				currentMetadata = newMeta
			}
			conflictDetect = 0
		}
		log.Printf("Metadata now is %+v (%f volume diff, %d conflicts, %d same, actual: %+v)", currentMetadata, maxVolumeDifference, conflictDetect, sameMeta, lastDetectedMeta)
	}
	getDevice := func(dev *stagelinq.Device) (devMeta map[int]*DeviceMeta) {
		devMeta, ok := metadata[dev]
		if !ok {
			devMeta = map[int]*DeviceMeta{}
			metadata[dev] = devMeta
		}
		return
	}
	getDeck := func(dev *stagelinq.Device, deckNum int) (deck *DeviceMeta) {
		device := getDevice(dev)
		deck, ok := device[deckNum]
		if !ok {
			deck = new(DeviceMeta)
			device[deckNum] = deck
		}
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			detectNewMetadata()
			go sendMetadata()
		case state := <-tracker.C():
			log.Printf("%s %s %+v", state.Device.Name, state.State.Name, state.State.Value)
			switch state.State.Name {
			case stagelinq.EngineDeck1TrackArtistName:
				getDeck(state.Device, 0).Artist = state.State.Value["string"].(string)
			case stagelinq.EngineDeck1TrackSongName:
				getDeck(state.Device, 0).Title = state.State.Value["string"].(string)
			case stagelinq.MixerCH1faderPosition:
				getDeck(state.Device, 0).Fader = state.State.Value["value"].(float64)
			case stagelinq.EngineDeck1Play:
				getDeck(state.Device, 0).Playing = state.State.Value["state"].(bool)
			case stagelinq.EngineDeck2TrackArtistName:
				getDeck(state.Device, 1).Artist = state.State.Value["string"].(string)
			case stagelinq.EngineDeck2TrackSongName:
				getDeck(state.Device, 1).Title = state.State.Value["string"].(string)
			case stagelinq.MixerCH2faderPosition:
				getDeck(state.Device, 1).Fader = state.State.Value["value"].(float64)
			case stagelinq.EngineDeck2Play:
				getDeck(state.Device, 1).Playing = state.State.Value["state"].(bool)
			case stagelinq.EngineDeck3TrackArtistName:
				getDeck(state.Device, 2).Artist = state.State.Value["string"].(string)
			case stagelinq.EngineDeck3TrackSongName:
				getDeck(state.Device, 2).Title = state.State.Value["string"].(string)
			case stagelinq.MixerCH3faderPosition:
				getDeck(state.Device, 2).Fader = state.State.Value["value"].(float64)
			case stagelinq.EngineDeck3Play:
				getDeck(state.Device, 2).Playing = state.State.Value["state"].(bool)
			case stagelinq.EngineDeck4TrackArtistName:
				getDeck(state.Device, 3).Artist = state.State.Value["string"].(string)
			case stagelinq.EngineDeck4TrackSongName:
				getDeck(state.Device, 3).Title = state.State.Value["string"].(string)
			case stagelinq.MixerCH4faderPosition:
				getDeck(state.Device, 3).Fader = state.State.Value["value"].(float64)
			case stagelinq.EngineDeck4Play:
				getDeck(state.Device, 3).Playing = state.State.Value["state"].(bool)
			}
		}
	}
}

type DeviceMeta struct {
	Playing bool
	Artist  string
	Title   string
	Fader   float64
}
