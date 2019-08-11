package main

import (
	"flag"
	"time"
	"strconv"
	"image"
	"image/draw"
	"github.com/Sunoo/go-rpi-rgb-led-matrix"
)

var (
	rows                     = flag.Int("led-rows", 16, "number of rows supported")
	cols                     = flag.Int("led-cols", 32, "number of columns supported")
	parallel                 = flag.Int("led-parallel", 1, "number of daisy-chained panels")
	chain                    = flag.Int("led-chain", 1, "number of displays daisy-chained")
	brightness               = flag.Int("brightness", 100, "brightness (0-100)")
	hardware_mapping         = flag.String("led-gpio-mapping", "regular", "Name of GPIO mapping used.")
	show_refresh             = flag.Bool("led-show-refresh", false, "Show refresh rate.")
	inverse_colors           = flag.Bool("led-inverse", false, "Switch if your matrix has inverse colors on.")
	disable_hardware_pulsing = flag.Bool("led-no-hardware-pulse", false, "Don't use hardware pin-pulse generation.")
)

const (
	/*xOrig = 1
	yOrig = 2
	segLength = 4*/
	xOrig = -6
	yOrig = 1
	segLength = 6
	leadingZero = false
	timeFormat = "03:04"
	animSpeed = 30
)

func main() {
	config := &rgbmatrix.DefaultConfig
	config.Rows = *rows
	config.Cols = *cols
	config.Parallel = *parallel
	config.ChainLength = *chain
	config.Brightness = *brightness
	config.HardwareMapping = *hardware_mapping
	config.ShowRefreshRate = *show_refresh
	config.InverseColors = *inverse_colors
	config.DisableHardwarePulsing = *disable_hardware_pulsing
	
	m, _ := rgbmatrix.NewRGBLedMatrix(config)

	canvas := rgbmatrix.NewCanvas(m)
	defer canvas.Close()
	
	d1 := NewDigit(segLength)
	d2 := NewDigit(segLength)
	co := NewColon(segLength)
	d3 := NewDigit(segLength)
	d4 := NewDigit(segLength)
	
	d1start := image.Point{xOrig, yOrig}
	d2start := d1start.Add(image.Point{d1.img.Bounds().Max.X + 1, 0})
	costart := d2start.Add(image.Point{d2.img.Bounds().Max.X + 1, 0})
	d3start := costart.Add(image.Point{co.img.Bounds().Max.X + 1, 0})
	d4start := d3start.Add(image.Point{d3.img.Bounds().Max.X + 1, 0})
	
	d1pos := image.Rectangle{d1start, d1start.Add(d1.img.Bounds().Max)}
	d2pos := image.Rectangle{d2start, d2start.Add(d2.img.Bounds().Max)}
	copos := image.Rectangle{costart, costart.Add(co.img.Bounds().Max)}
	d3pos := image.Rectangle{d3start, d3start.Add(d3.img.Bounds().Max)}
	d4pos := image.Rectangle{d4start, d4start.Add(d4.img.Bounds().Max)}
	
	initialTime := true
	
	for {
		clock := time.Now().Format(timeFormat)
		h1, _ := strconv.Atoi(clock[0:1])
		h2, _ := strconv.Atoi(clock[1:2])
		m1, _ := strconv.Atoi(clock[3:4])
		m2, _ := strconv.Atoi(clock[4:5])

		if initialTime {
			draw.Draw(canvas, copos, co.img, image.ZP, draw.Src)
			d4.Draw(m2)
			draw.Draw(canvas, d4pos, d4.img, image.ZP, draw.Src)
			d3.Draw(m1)
			draw.Draw(canvas, d3pos, d3.img, image.ZP, draw.Src)
			d2.Draw(h2)
			draw.Draw(canvas, d2pos, d2.img, image.ZP, draw.Src)
			if leadingZero || h1 != 0 {
				d1.Draw(h1)
				draw.Draw(canvas, d1pos, d1.img, image.ZP, draw.Src)
			}
		
			canvas.RenderKeep()
			initialTime = false
		} else {
			if m2 != d4.value {
				done := false
				for !done {
					done = d4.Morph(m2)
					draw.Draw(canvas, d4pos, d4.img, image.ZP, draw.Src)
					canvas.RenderKeep()
					time.Sleep(animSpeed * time.Millisecond)
				}
			}
			if m1 != d3.value {
				done := false
				for !done {
					done = d3.Morph(m1)
					draw.Draw(canvas, d3pos, d3.img, image.ZP, draw.Src)
					canvas.RenderKeep()
					time.Sleep(animSpeed * time.Millisecond)
				}
			}
			if h2 != d2.value {
				done := false
				for !done {
					done = d2.Morph(h2)
					draw.Draw(canvas, d2pos, d2.img, image.ZP, draw.Src)
					canvas.RenderKeep()
					time.Sleep(animSpeed * time.Millisecond)
				}
			}
			if h1 != d1.value {
				done := false
				for !done {
					done = d1.Morph(h1)
					draw.Draw(canvas, d1pos, d1.img, image.ZP, draw.Src)
					canvas.RenderKeep()
					time.Sleep(animSpeed * time.Millisecond)
				}
			}
		}
		target := time.Now().Truncate(time.Minute).Add(time.Minute)
		time.Sleep(target.Sub(time.Now()))
	}
}