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
//	"./hough"
//	"./edges"
//	"./geometry"
)

var window Window

var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")

func main() {
	file,err := os.Open("fig1.png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer file.Close()

	img,err := png.Decode(file)
	window.Bg = img
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	flag.Parse()
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *Circle, 400)

	// Create X11 window
	xwindow,err := x11.NewWindow()
	window.Window = xwindow
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	window.Screen = xwindow.Screen()
	redraw(window)
	go events(xwindow.EventChan(), window)

	start := time.Seconds()

	r := findBounds(&img)

	mid := time.Seconds()

	go filter(&img,filtered)
	go Sobel(&img,filtered,edge)
	go Transform(&img, edge, transformed, r)

	bx,by := img.Bounds().Max.X,img.Bounds().Max.Y
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
			window.Centre = *c
		}
	}

	DrawCircle(window.Screen, window.Centre.Centre, window.Centre.Radius, image.RGBAColor{0,255,255,255})
	window.Window.FlushImage()

	end := time.Seconds()
	fmt.Printf("Done in %d seconds, findBounds() took %d seconds\n", end-start, mid-start)

	for {
	}
}
