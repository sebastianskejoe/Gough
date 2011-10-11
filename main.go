package main

import (
	"fmt"
	"image"
	"time"
	"os"
	"flag"
	"runtime/pprof"
	"exp/gui/x11"
	"exp/gui"
	"strings"
	"strconv"
)

var window Window

var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")
var calPath = flag.String("cal", "fig1.png", "The calibration image")
var imgPath = flag.String("img", "", "The second image")
var dist	= flag.Int("dist", 0, "Distance from camera to trackpoint")
var width	= flag.Int("width", 0, "The diameter of the trackpoint in centimeters")


func findCircle(img *image.Image) Circle {
	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *Circle, 400)
	var centre Circle


	fmt.Printf("Finding circle\n")

	window.State = WORKING
	redraw(&window)

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

	window.State = IDLE

	return centre
}

func TrimFunc(c int) bool {
	if c == '[' || c == ']' {
		return false
	}
	return true
}

func main() {
	flag.Parse()

	// Parse image paths
	fmt.Printf("%s\n", *imgPath)
	if *imgPath != "" {
		r := strings.TrimFunc(*imgPath, TrimFunc)
		split := strings.Split(r[1:len(r)-1], "-")
		start,_ := strconv.Atoi(split[0])
		end,_ := strconv.Atoi(split[1])
		f := 0
		for i := start ; i <= end ; i++ {
			window.Frames = append(window.Frames, Frame{Id: f, Path: strings.Replace(*imgPath, r, strconv.Itoa(i),-1)})
			f++
		}
		window.FrameCount = f
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

	// Calibration data
	window.Dist = *dist

	// Create X11 window
	xwindow,err := x11.NewWindow()
	window.Window = xwindow
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	window.Screen = xwindow.Screen()
	redraw(&window)


	// Event loop
	for {
		event, ok := <-xwindow.EventChan()
		if ok != true {
			break
		}

		switch e := event.(type) {
		case gui.ConfigEvent:
			window.Screen = xwindow.Screen()
			redraw(&window)
			break
		case gui.MouseEvent:
			redraw(&window)
			break
		case gui.KeyEvent:
			fmt.Printf("Key: %d\n", e.Key)
			switch e.Key {
			case 99:
				fmt.Println("Starting calibration")
				// Find circle in calibration image
/*				go func() {
					c := findCircle(&img)
					window.Centre = c
					if *width != 0 {
						window.Ppc = (window.Centre.Radius*2)/(*width)
					}
					redraw(&window)
				} ()*/
				break
			case 98:
				go func() {
					cfra := window.Cfra
					for ; window.State == WORKING ; {
						fmt.Printf("Waiting %d\n", window.State)
						time.Sleep(1e9)
					}

					window.Frames[cfra].Centre = findCircle(&window.Frames[cfra].img)
					redraw(&window)
				} ()
			case 65361:
				if window.Cfra > 0 {
					window.Cfra--
				}
				redraw(&window)
			case 65363:
				if window.Cfra < window.FrameCount-1 {
					window.Cfra++
				}
				redraw(&window)
				break
			default:
				break
			}
		}
	}
}
