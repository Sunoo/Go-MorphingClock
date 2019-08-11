package main

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	sA int = 0
	sB int = 1
	sC int = 2
	sD int = 3
	sE int = 4
	sF int = 5
	sG int = 6
)

var (
	digitBits = [10][7]bool {{true, true, true, true, true, true, false},
		{false, true, true, false, false, false, false},
		{true, true, false, true, true, false, true},
		{true, true, true, true, false, false, true},
		{false, true, true, false, false, true, true},
		{true, false, true, true, false, true, true},
		{true, false, true, true, true, true, true},
		{true, true, true, false, false, false, false},
		{true, true, true, true, true, true, true},
		{true, true, true, true, false, true, true}}
)

type Digit struct {
	segLength int
	value int
	img draw.Image
	frame int
}

func NewDigit(segLength int) Digit {
	width := segLength + 2
	height := segLength * 2 + 3
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	d := Digit{segLength, -1, img, 0}
	return d
}

func NewColon(segLength int) Digit {
	width := 2
	height := segLength * 2 + 3
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	d := Digit{segLength, -1, img, 0}
	//For some reason I can't find, the height needs to be 1 instead of 2 to be 2 pixels tall
	d.drawRect(0, segLength + 3, 2, 1, true)
	d.drawRect(0, segLength + 1 - 3, 2, 1, true)
	return d
}

func (d *Digit) drawPixel(x int, y int, active bool) {
	var col color.RGBA
	if active {
		col = color.RGBA{0, 0, 255, 255}
	} else {
		col = color.RGBA{0, 0, 0, 255}
	}
	d.img.Set(x, d.img.Bounds().Max.Y - y - 1, col)
}

func (d *Digit) drawRect(x int, y int, w int, h int, active bool) {
	for curX := x; curX <= x + w; curX++ {
		for curY := y; curY <= y + h; curY++ {
			d.drawPixel(curX, curY, active)
		}
	}
}

func (d *Digit) drawLine(x1 int, y1 int, x2 int, y2 int, active bool) {
	var x, y, w, h int
	if x1 < x2 {
		x = x1
		w = x2 - x1
	} else {
		x = x2
		w = x1 - x2
	}
	if y1 < y2 {
		y = y1
		h = y2 - y1
	} else {
		y = y2
		h = y1 - y2
	}
	d.drawRect(x, y, w, h, active)
}

func (d *Digit) drawSeg(seg int) {
	switch seg {
		case sA:
			d.drawLine(1, d.segLength * 2 + 2, d.segLength, d.segLength * 2 + 2, true)
		case sB:
			d.drawLine(d.segLength + 1, d.segLength * 2 + 1, d.segLength + 1, d.segLength + 2, true)
		case sC:
			d.drawLine(d.segLength + 1, 1, d.segLength + 1, d.segLength, true)
		case sD:
			d.drawLine(1, 0, d.segLength, 0, true)
		case sE:
			d.drawLine(0, 1, 0, d.segLength, true)
		case sF:
			d.drawLine(0, d.segLength * 2 + 1, 0, d.segLength + 2, true)
		case sG:
			d.drawLine(1, d.segLength + 1, d.segLength, d.segLength + 1, true)
	}
}

func (d *Digit) Draw(value int) {
	if value == -1 {
		draw.Draw(d.img, d.img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	} else {
		for i := 0; i < 7; i++ {
			if digitBits[value][i] {d.drawSeg(i)}
		}
	}
	d.value = value
	d.frame = 0
}

func (d *Digit) Blank() {
	d.Draw(-1)
}

func (d *Digit) Morph1() bool {
	if d.frame <= d.segLength + 1 {
		d.drawLine(d.frame - 1, 1, d.frame - 1, d.segLength, false)
		d.drawLine(d.frame, 1, d.frame, d.segLength, true)
		
		d.drawLine(d.frame - 1, d.segLength * 2 + 1, d.frame - 1, d.segLength + 2, false)
		d.drawLine(d.frame, d.segLength * 2 + 1, d.frame, d.segLength + 2, true)
		
		d.drawPixel(1 + d.frame, d.segLength * 2 + 2, false)
		d.drawPixel(1 + d.frame, 0, false)
		d.drawPixel(1 + d.frame, d.segLength + 1, false)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph2() bool {
	if d.frame <= d.segLength {
		if (d.frame < d.segLength) {
			d.drawPixel(d.segLength - d.frame, d.segLength * 2 + 2, true)
			d.drawPixel(d.segLength - d.frame, d.segLength + 1, true)
			d.drawPixel(d.segLength - d.frame, 0, true)
		}
		
		d.drawLine(d.segLength + 1 - d.frame, 1, d.segLength + 1 - d.frame, d.segLength, false)
		d.drawLine(d.segLength - d.frame, 1, d.segLength - d.frame, d.segLength, true)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph3() bool {
	if d.frame <= d.segLength {
		d.drawLine(d.frame, 1, d.frame, d.segLength, false)
		d.drawLine(1 + d.frame, 1, 1 + d.frame, d.segLength, true)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph4() bool {
	if d.frame < d.segLength {
		d.drawPixel(d.segLength - d.frame, d.segLength * 2 + 2, false)
		d.drawPixel(0, d.segLength * 2 + 1 - d.frame, true)
		d.drawPixel(1 + d.frame, 0, false)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph5() bool {
	if d.frame < d.segLength {
		d.drawPixel(d.segLength + 1, d.segLength + 2 + d.frame, false)
		d.drawPixel(d.segLength - d.frame, d.segLength * 2 + 2, true)
		d.drawPixel(d.segLength - d.frame, 0, true)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph6() bool {
	if d.frame <= d.segLength {
		d.drawLine(d.segLength - d.frame, 1, d.segLength - d.frame, d.segLength, true)
		if d.frame > 0 {
			d.drawLine(d.segLength - d.frame + 1, 1, d.segLength - d.frame + 1, d.segLength, false)
		}
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph7() bool {
	if d.frame <= d.segLength + 1 {
		d.drawLine(d.frame - 1, 1, d.frame - 1, d.segLength, false)
		d.drawLine(d.frame, 1, d.frame, d.segLength, true)
		
		d.drawLine(d.frame - 1, d.segLength * 2 + 1, d.frame - 1, d.segLength + 2, false)
		d.drawLine(d.frame, d.segLength * 2 + 1, d.frame, d.segLength + 2, true)
		
		d.drawPixel(1 + d.frame, 0, false)
		d.drawPixel(1 + d.frame, d.segLength + 1, false)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph8() bool {
	if d.frame <= d.segLength {
		d.drawLine(d.segLength - d.frame, d.segLength * 2 + 1, d.segLength - d.frame, d.segLength + 2, true)
		if d.frame > 0 {
			d.drawLine(d.segLength - d.frame + 1, d.segLength * 2 + 1, d.segLength - d.frame + 1, d.segLength + 2, false)
		}
		
		d.drawLine(d.segLength - d.frame, 1, d.segLength - d.frame, d.segLength, true)
		if d.frame > 0 {
			d.drawLine(d.segLength - d.frame + 1, 1, d.segLength - d.frame + 1, d.segLength, false)
		}
		
		if d.frame < d.segLength {
			d.drawPixel(d.segLength - d.frame, 0, true)
			d.drawPixel(d.segLength - d.frame, d.segLength + 1, true)
		}
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph9() bool {
	if d.frame <= d.segLength + 1 {
		d.drawLine(d.frame - 1, 1, d.frame - 1, d.segLength, false)
		d.drawLine(d.frame, 1, d.frame, d.segLength, true)
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph0() bool {
	if d.frame <= d.segLength {
		if d.value == 1 {
			d.drawLine(d.segLength - d.frame, d.segLength * 2 + 1, d.segLength - d.frame, d.segLength + 2, true)
			if d.frame > 0 {
				d.drawLine(d.segLength - d.frame + 1, d.segLength * 2 + 1, d.segLength - d.frame + 1, d.segLength + 2, false)
			}
			
			d.drawLine(d.segLength - d.frame, 1, d.segLength - d.frame, d.segLength, true)
			if d.frame > 0 {
				d.drawLine(d.segLength - d.frame + 1, 1, d.segLength - d.frame + 1, d.segLength, false)
			}
			
			if d.frame < d.segLength {
				d.drawPixel(d.segLength - d.frame, d.segLength * 2 + 2, true)
				d.drawPixel(d.segLength - d.frame, 0, true)
			}
		}
		
		if d.value == 2 {
			d.drawLine(d.segLength - d.frame, d.segLength * 2 + 1, d.segLength - d.frame, d.segLength + 2, true)
			if d.frame > 0 {
				d.drawLine(d.segLength - d.frame + 1, d.segLength * 2 + 1, d.segLength - d.frame + 1, d.segLength + 2, false)
			}
			
			d.drawPixel(1 + d.frame, d.segLength + 1, false)
			if d.frame < d.segLength {
				d.drawPixel(d.segLength + 1, d.segLength + 1 - d.frame, true)
			}
		}
		
		if d.value == 5 {
			if d.frame < d.segLength {
				if d.frame > 0 {
					d.drawLine(1 + d.frame, d.segLength * 2 + 1, 1 + d.frame, d.segLength + 2, false)
				}
				d.drawLine(2 + d.frame, d.segLength * 2 + 1, 2 + d.frame, d.segLength + 2, true)
			}
		}
		
		if d.value == 5 || d.value == 9 {
			if d.frame < d.segLength {
				d.drawPixel(d.segLength - d.frame, d.segLength + 1, false)
				d.drawPixel(0, d.segLength - d.frame, true)
			}
		}
		
		d.frame++
		return false
	} else {
		d.frame = 0
		return true
	}
}

func (d *Digit) Morph(value int) bool {
	if d.value == -1 {
		d.Draw(value)
		return true
	} else {
	done := false
	switch value {
		case 1:
			done = d.Morph1()
		case 2:
			done = d.Morph2()
		case 3:
			done = d.Morph3()
		case 4:
			done = d.Morph4()
		case 5:
			done = d.Morph5()
		case 6:
			done = d.Morph6()
		case 7:
			done = d.Morph7()
		case 8:
			done = d.Morph8()
		case 9:
			done = d.Morph9()
		case 0:
			done = d.Morph0()
		case -1:
			d.Blank()
			done = true
	}
	if done {
		d.value = value
	}
	return done
	}
}