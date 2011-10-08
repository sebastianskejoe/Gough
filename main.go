package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"
	"flag"
	"runtime/pprof"
	"exp/gui/x11"
	"exp/gui"
)

var window Window

var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")
var imgPath = flag.String("img", "fig1.png", "The image")
var dist	= flag.Int("dist", 0, "Distance from camera to trackpoint")
var width	= flag.Int("width", 0, "The diameter of the trackpoint in centimeters")


func findCircle(img *image.Image) Circle {
	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *Circle, 400)
	var centre Circle

	start := time.Seconds()

	r := findBounds(img)

	mid := time.Seconds()

	go filter(img,filtered)
	go Sobel(img,filtered,edge)
	go Transform(img, edge, transformed, r)

	bx,by := (*img).Bounds().Max.X,(*img).Bounds().Max.Y
	votes := make([]int, bx*by*bx)
	max := 0
	for {
		c,ok1 := <-transformed

		if ok1 != true {
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
	fmt.Printf("Done in %d seconds, findBounds() took %d seconds\n", end-start, mid-start)

	return centre
}

func main() {
	flag.Parse()

	// Get image
	file,err := os.Open(*imgPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer file.Close()

	img,err := png.Decode(file)
	window.Bg = img
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Do cpu profiling if asked for
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	// Create X11 window
	xwindow,err := x11.NewWindow()
	window.Window = xwindow
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	window.Screen = xwindow.Screen()
	redraw(&window)

	// Do hough transformation

	go func() {
		c := findCircle(&img)
		window.Centre = c
		redraw(&window)
	} ()

	// Event loop
	for {
		event, ok := <-xwindow.EventChan()
		if ok != true {
			break
		}

		switch e := event.(type) {
		case gui.ConfigEvent:
			redraw(&window)
			break
		case gui.MouseEvent:
			redraw(&window)
			break
		}
	}
}
