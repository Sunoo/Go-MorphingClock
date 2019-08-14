package main

import (
	"time"
	"strconv"
	"os"
	"net"
	"bytes"
	"image"
	"image/draw"
	"image/color"
	"encoding/json"
	"io/ioutil"
	"github.com/Sunoo/go-rpi-rgb-led-matrix"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/Sunoo/hsv"
	"github.com/lmittmann/ppm"
)

var (
	clockConfig = ClockConfig{hsv.HSVColor{240, 100, 30}, 1, 2, 4, false, "03:04", 30 * time.Millisecond, time.Second / 2, 16, 32, 1, 1, "regular", false, false, false}
	matrix rgbmatrix.Matrix
	canvas *rgbmatrix.Canvas
	sketch *image.RGBA
	power = true
	stopchan chan bool
	stoppedchan chan bool
)

type ClockConfig struct {
	ClockColor hsv.HSVColor
	XOrig int
	YOrig int
	SegLength int
	LeadingZero bool
	TimeFormat string
	AnimSpeed time.Duration
	ClockReturn time.Duration
	Rows int
	Cols int
	Parallel int
	ChainLength int
	HardwareMapping string
	ShowRefreshRate bool
	InverseColors bool
	DisableHardwarePulsing bool
}

func Render() {
	if (power) {
		newR, newG, newB, _ := clockConfig.ClockColor.RGBA()
		bounds := sketch.Bounds()
		for curX := bounds.Min.X; curX < bounds.Max.X; curX++ {
			for curY := bounds.Min.Y; curY < bounds.Max.Y; curY++ {
				curR, curG, curB, curA := sketch.At(curX, curY).RGBA()
				curR = curR * newR / 255
				curG = curG * newG / 255
				curB = curB * newB / 255
				canvas.Set(curX, curY, color.RGBA{uint8(curR), uint8(curG), uint8(curB), uint8(curA)})
			}
		}
		canvas.Render()
	}
}

func RunClock() {
	sketch = image.NewRGBA(canvas.Bounds())
	
	d1 := NewDigit(clockConfig.SegLength)
	d2 := NewDigit(clockConfig.SegLength)
	co := NewColon(clockConfig.SegLength)
	d3 := NewDigit(clockConfig.SegLength)
	d4 := NewDigit(clockConfig.SegLength)
	
	d1start := image.Point{clockConfig.XOrig, clockConfig.YOrig}
	d2start := d1start.Add(image.Point{d1.img.Bounds().Max.X + 1, 0})
	costart := d2start.Add(image.Point{d2.img.Bounds().Max.X + 1, 0})
	d3start := costart.Add(image.Point{co.img.Bounds().Max.X + 1, 0})
	d4start := d3start.Add(image.Point{d3.img.Bounds().Max.X + 1, 0})
	
	d1pos := image.Rectangle{d1start, d1start.Add(d1.img.Bounds().Max)}
	d2pos := image.Rectangle{d2start, d2start.Add(d2.img.Bounds().Max)}
	copos := image.Rectangle{costart, costart.Add(co.img.Bounds().Max)}
	d3pos := image.Rectangle{d3start, d3start.Add(d3.img.Bounds().Max)}
	d4pos := image.Rectangle{d4start, d4start.Add(d4.img.Bounds().Max)}
	
	warmUp := true
	
	for {
		clock := time.Now().Format(clockConfig.TimeFormat)
		h1, _ := strconv.Atoi(clock[0:1])
		h2, _ := strconv.Atoi(clock[1:2])
		m1, _ := strconv.Atoi(clock[3:4])
		m2, _ := strconv.Atoi(clock[4:5])
		
		if !clockConfig.LeadingZero && h1 == 0 {
			h1 = -1
		}

		if m2 != d4.value {
			done := false
			for !done {
				done = d4.Morph(m2)
				draw.Draw(sketch, d4pos, d4.img, image.ZP, draw.Src)
				Render()
				time.Sleep(clockConfig.AnimSpeed)
			}
		}
		if m1 != d3.value {
			done := false
			for !done {
				done = d3.Morph(m1)
				draw.Draw(sketch, d3pos, d3.img, image.ZP, draw.Src)
				Render()
				time.Sleep(clockConfig.AnimSpeed)
			}
		}
		if warmUp {
			draw.Draw(sketch, copos, co.img, image.ZP, draw.Src)
			Render()
			time.Sleep(clockConfig.AnimSpeed)
		}
		if h2 != d2.value {
			done := false
			for !done {
				done = d2.Morph(h2)
				draw.Draw(sketch, d2pos, d2.img, image.ZP, draw.Src)
				Render()
				time.Sleep(clockConfig.AnimSpeed)
			}
		}
		if h1 != d1.value {
			done := false
			for !done {
				done = d1.Morph(h1)
				draw.Draw(sketch, d1pos, d1.img, image.ZP, draw.Src)
				Render()
				time.Sleep(clockConfig.AnimSpeed)
			}
		}
		select {
			case <-time.After(time.Now().Truncate(time.Minute).Add(time.Minute).Sub(time.Now())):
				//Just keep running
			case <-stopchan:
				canvas.Render()
				stoppedchan <- true
				return
		}
	}
}

func Flaschen() {
	pc, err := net.ListenPacket("udp", ":1337")
	if err != nil {
		return
	}
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, 65535)
	clockStopped := false

	f := func() {
		clockStopped = false
		matrix.SetBrightness(clockConfig.ClockColor.V)
		go RunClock()
	}
	timer := time.AfterFunc(clockConfig.ClockReturn, f)
	timer.Stop()

	go func() {
		for {
			n, _, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}
			
			if !clockStopped {
				matrix.SetBrightness(100)
				stopchan <- true
				<-stoppedchan
				clockStopped = true
			}
			
			timer.Reset(clockConfig.ClockReturn)
			
			img, err := ppm.Decode(bytes.NewReader(buffer[:n]))
			if err != nil {
				doneChan <- err
				return
			}
			
			draw.Draw(canvas, canvas.Bounds(), img, image.ZP, draw.Src)
    		canvas.Render()
		}
	}()

	select {}
}

func main() {
	jsonConfig, _ := ioutil.ReadFile("config.json")
	json.Unmarshal(jsonConfig, &clockConfig)
	
	stopchan = make(chan bool)
	stoppedchan = make(chan bool)
	
	config := &rgbmatrix.DefaultConfig
	config.Rows = clockConfig.Rows
	config.Cols = clockConfig.Cols
	config.Parallel = clockConfig.Parallel
	config.ChainLength = clockConfig.ChainLength
	config.Brightness = clockConfig.ClockColor.V
	config.HardwareMapping = clockConfig.HardwareMapping
	config.ShowRefreshRate = clockConfig.ShowRefreshRate
	config.InverseColors = clockConfig.InverseColors
	config.DisableHardwarePulsing = clockConfig.DisableHardwarePulsing
	
	matrix, _ = rgbmatrix.NewRGBLedMatrix(config)

	canvas = rgbmatrix.NewCanvas(matrix)
	defer canvas.Close()
	
	info := accessory.Info{
		Name:         "Clock",
		SerialNumber: "rpi-rgb-led-matrix",
		Manufacturer: "Sunoo",
		Model:        "Morphing Clock",
	}

	acc := accessory.NewLightbulb(info)
	
	acc.Lightbulb.On.SetValue(true)
	acc.Lightbulb.Brightness.SetValue(clockConfig.ClockColor.V)
	acc.Lightbulb.Saturation.SetValue(clockConfig.ClockColor.S)
	acc.Lightbulb.Hue.SetValue(clockConfig.ClockColor.H)
	
	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if power != on {
			power = on;
			if on {
				go RunClock()
			} else {
				stopchan <- true
				<-stoppedchan
			}
		}
	})
	
	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(bright int) {
		clockConfig.ClockColor.V = bright
		matrix.SetBrightness(bright)
		Render()
	})
	
	acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(sat float64) {
		clockConfig.ClockColor.S = sat
		Render()
	})
	
	acc.Lightbulb.Hue.OnValueRemoteUpdate(func(hue float64) {
		clockConfig.ClockColor.H = hue
		Render()
	})

	t, _ := hc.NewIPTransport(hc.Config{Pin: "12312312"}, acc.Accessory)
	
	hc.OnTermination(func() {
		<-t.Stop()
		jsonConfig, _ := json.MarshalIndent(clockConfig, "", "\t")
		ioutil.WriteFile("config.json", jsonConfig, 0666)
		os.Exit(0)
	})

	go t.Start()
	
	go RunClock()
	
	go Flaschen()
	
	select {}
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}