package main

import (
	"fmt"
	"time"
	"os"
	"flag"
	"runtime/pprof"
	"exp/gui/x11"
	"exp/gui"
	"strings"
	"strconv"
)


var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")
var calPath = flag.String("cal", "fig1.png", "The calibration image")
var imgPath = flag.String("img", "", "The second image")
var dist	= flag.Int("dist", 0, "Distance from camera to trackpoint")
var width	= flag.Int("width", 0, "The diameter of the trackpoint in centimeters")

func main() {
	var window Window
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
			path := strings.Replace(*imgPath, r, strconv.Itoa(i),-1)
			window.Frames = append(window.Frames, Frame{Id: f, Path: path})
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
			case 'c':
			case KEY_SPACE: // space
				go func() {
					cfra := window.Cfra

					for ; GetState(&window) == WORKING ; {
						time.Sleep(1e9)
					}

					window.Frames[cfra].Centre = findCircle(&window,&window.Frames[cfra].img)
					redraw(&window)
				} ()
			case KEY_LEFT:
				if window.Cfra > 0 {
					window.Cfra--
				}
				redraw(&window)
			case KEY_RIGHT:
				if window.Cfra < window.FrameCount-1 {
					window.Cfra++
				}
				redraw(&window)
			default:
			}
		}
	}
}
