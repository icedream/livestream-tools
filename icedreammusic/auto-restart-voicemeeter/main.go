package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	asio "github.com/JamesDunne/go-asio"
	"github.com/cenkalti/backoff/v4"
)

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func main() {
	// TODO - handle interrupt signal here rather than in sample loop

	bo := backoff.NewConstantBackOff(time.Second)
	err := backoff.Retry(run, bo)
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Printf("CoInitialize(0)\n")
	asio.CoInitialize(0)
	defer fmt.Printf("CoUninitialize()\n")
	defer asio.CoUninitialize()

	drivers, err := asio.ListDrivers()
	if err != nil {
		return backoff.Permanent(err)
	}

	var mainOutDriver *asio.ASIODriver
	for _, driver := range drivers {
		if driver.Name != "Voicemeeter Virtual ASIO" {
			continue
		}
		mainOutDriver = driver
		break
	}

	if mainOutDriver == nil {
		return backoff.Permanent(errors.New("could not find main Voicemeeter ASIO output"))
	}

	log.Println("ASIO driver:", mainOutDriver.GUID, mainOutDriver.CLSID, mainOutDriver.Name)

	if err := mainOutDriver.Open(); err != nil {
		return err
	}
	defer mainOutDriver.Close()

	drv := mainOutDriver.ASIO
	fmt.Printf("getDriverName():      '%s'\n", drv.GetDriverName())
	fmt.Printf("getDriverVersion():   %d\n", drv.GetDriverVersion())

	// mainOutUnknown := drv.AsIUnknown()
	// mainOutUnknown.AddRef()
	// defer mainOutUnknown.Release()

	// getChannels
	in, out, err := drv.GetChannels()
	if err != nil {
		return err
	}
	fmt.Printf("getChannels():        %d, %d\n", in, out)

	// getBufferSize
	minSize, maxSize, preferredSize, granularity, err := drv.GetBufferSize()
	if err != nil {
		return err
	}
	fmt.Printf("getBufferSize():      %d, %d, %d, %d\n", minSize, maxSize, preferredSize, granularity)

	// getSampleRate
	srate, err := drv.GetSampleRate()
	if err != nil {
		return err
	}
	fmt.Printf("getSampleRate():      %v\n", srate)

	// canSampleRate
	var sampleRate float64
	// for _, canSampleRate := range []int{
	// 	48000,
	// 	44100,
	// } {
	// 	sampleRateF64 := float64(canSampleRate)
	// 	err = drv.CanSampleRate(sampleRateF64)
	// 	fmt.Printf("canSampleRate(%q): %v\n", sampleRateF64, err)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	sampleRate = canSampleRate
	// 	break
	// }
	// if sampleRate == 0 {
	// 	// log.Fatal("Could not negotiate a compatible samplerate")
	// 	// return
	// 	fmt.Println("WARNING: Defaulting to 48000 Hz")
	// 	sampleRate = 48000
	// }
	sampleRate = srate
	// sampleRate = 48000

	// SetSampleRate
	err = drv.SetSampleRate(sampleRate)
	fmt.Printf("setSampleRate(%v): %v\n", float64(sampleRate), err)
	if err != nil {
		fmt.Println("WARNING: setSampleRate failed, ignoring:", err)
	}

	// outputReady
	fmt.Printf("outputReady():        %v\n", drv.OutputReady())

	// open control panel:
	// drv.ControlPanel()

	bufferDescriptors := make([]asio.BufferInfo, 0, in+out)
	for i := 0; i < in; i++ {
		bufferDescriptors = append(bufferDescriptors, asio.BufferInfo{
			Channel: i,
			IsInput: true,
		})
		cinfo, err := drv.GetChannelInfo(i, true)
		if err != nil {
			log.Fatal(err)
			continue
		}
		fmt.Printf(" IN%-2d: active=%v, group=%d, type=%d, name=%s\n", i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
	}
	for i := 0; i < out; i++ {
		bufferDescriptors = append(bufferDescriptors, asio.BufferInfo{
			Channel: i,
			IsInput: false,
		})
		cinfo, err := drv.GetChannelInfo(i, false)
		if err != nil {
			log.Fatal(err)
			continue
		}
		fmt.Printf("OUT%-2d: active=%v, group=%d, type=%d, name=%s\n", i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
	}

	err = drv.CreateBuffers(bufferDescriptors, 512, asio.Callbacks{
		Message: func(selector, value int32, message uintptr, opt *float64) int32 {
			log.Println("Message:", selector, value, message, opt)
			return 0
		},
		BufferSwitch: func(doubleBufferIndex int, directProcess bool) {
			log.Println("Buffer switch:", doubleBufferIndex, directProcess)
		},
		BufferSwitchTimeInfo: func(params *asio.ASIOTime, doubleBufferIndex int32, directProcess bool) *asio.ASIOTime {
			log.Println("Buffer switch time info:", params, doubleBufferIndex, directProcess)
			return params
		},
		SampleRateDidChange: func(rate float64) {
			log.Println("Sample rate did change:", rate)
		},
	})
	if err != nil {
		return err
	}
	defer fmt.Printf("disposeBuffers()\n")
	defer drv.DisposeBuffers()
	fmt.Printf("createBuffers()\n")

	// getLatencies
	latin, latout, err := drv.GetLatencies()
	if err != nil {
		return err
	}
	fmt.Printf("getLatencies():       %d, %d\n", latin, latout)

	err = drv.Start()
	if err != nil {
		return err
	}
	defer drv.Stop()

	c := make(chan os.Signal, 1)
	go signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	output := new(strings.Builder)
	ladder := " ▁▂▃▄▅▆▇█"
	chars := []rune(reverse(ladder) + ladder)
	grace := 5 * time.Second
	lastSignalTime := time.Now()
	for {
		select {
		case <-time.After(33 * time.Millisecond):
			output.Reset()
			for _, desc := range bufferDescriptors {
				for _, buf := range desc.Buffers {
					output.WriteString("|")
					if buf == nil {
						output.WriteString("?")
						continue
					}
					if *buf != 0 {
						lastSignalTime = time.Now()
					}
					output.WriteString(fmt.Sprintf("%c",
						chars[len(chars)/2+
							int(float64(len(chars))/2*
								float64(*buf)/float64(math.MaxInt32))]))
				}
			}
			os.Stdout.WriteString(output.String() + "\r")
			if time.Now().Sub(lastSignalTime) > time.Second {
				os.Stdout.WriteString("Silence!\r")
			}
			if time.Now().Sub(lastSignalTime) > grace {
				os.Stdout.WriteString("\n")
				log.Println("Restarting audio engine...")
				cmd := exec.Command(`C:\Program Files (x86)\VB\Voicemeeter\voicemeeterpro.exe`, "-r")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				log.Printf("Restart audio engine result: %v", cmd.Run())
				// time.Sleep(3 * time.Second)
				// lastSignalTime = time.Now()
				return errors.New("audio engine restarted, retry")
			}
		case <-c:
			return nil
		}
	}
}
