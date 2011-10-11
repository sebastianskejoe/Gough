package main

import (
	"image"
	"fmt"
	"exp/gui"
	"image/draw"
	"os"
	"image/png"
	"sync"
	"time"
)

const (
	IDLE = iota
	WORKING
)

type Window struct {
	Screen draw.Image
	Window gui.Window
	Ppc int // pixel per centimeter
	Dist int // distance from camera to track point
	Cfra int // current frame
	State int
	FrameCount int
	Frames []Frame
	Calibration Frame
}

type Frame struct {
	Path string
	Id int
	Centre Circle
	img image.Image
}

func filter(img *image.Image, filtered chan<- *image.Point) {
	bx := (*img).Bounds().Max.X
	by := (*img).Bounds().Max.Y
	sent := 0
	for x := 0 ; x < bx ; x++ {
		for y := 0 ; y < by ; y++ {
			if ColorIsGood((*img).At(x,y)) {
				filtered <- &image.Point{x,y}
				sent++
			}
		}
	}
	fmt.Printf("Sent %d times in filter()\n", sent)
	close(filtered)
}

func ColorIsGood(c image.Color) bool {
	r,_,_,_ := c.RGBA()
	if r > 200 {
		return true
	}
	return false
}

func findBounds(img *image.Image) image.Rectangle {
	bx := (*img).Bounds().Max.X
	by := (*img).Bounds().Max.Y
	maxx,maxy,minx,miny := 0,0,bx,by

	for x := 0 ; x < bx ; x++ {
		for y := 0 ; y < by ; y++ {
			if ColorIsGood((*img).At(x,y)) {
				if x > maxx { maxx = x }
				if x < minx { minx = x }
				if y > maxy { maxy = y }
				if y < miny { miny = y }
			}
		}
	}

	return image.Rect(minx,miny,maxx,maxy)
}

func getImage(path string) image.Image {
	// Get image
	file,err := os.Open(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer file.Close()

	img,err := png.Decode(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	return img
}

func TrimFunc(c int) bool {
	if c == '[' || c == ']' {
		return false
	}
	return true
}

var statemutex sync.Mutex

func SetState(window *Window, state int) {
	statemutex.Lock()
	window.State = state
	statemutex.Unlock()
}

func GetState(window *Window) int {
	statemutex.Lock()
	defer statemutex.Unlock()
	return window.State
}

func findCircle(window *Window, img *image.Image) Circle {
	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *Circle, 400)
	var centre Circle


	fmt.Printf("Finding circle\n")

	SetState(window, WORKING)
	redraw(window)

	start := time.Seconds()

	r := findBounds(img)

	go filter(img,filtered)
	go Sobel(img,filtered,edge)
	go Transform(img, edge, transformed, r)

	bx,by := (*img).Bounds().Max.X,(*img).Bounds().Max.Y
	votes := make([]int, bx*by*bx)
	max := 0
	for {
		c,ok := <-transformed

		if ok != true {
			break
		}

		index := (c.Radius-1)+bx*c.Centre.Y+bx*by*c.Centre.X
		votes[index] += 1
		if votes[index] > max {
			max = votes[index]
			centre = *c
		}
	}
	end := time.Seconds()
	fmt.Printf("Done in %d seconds\n", end-start)

	SetState(window, IDLE)

	return centre
}
