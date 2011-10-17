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
	"runtime"
)



var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var memprof = flag.String("memprofile", "", "Memory profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")
var calPath = flag.String("cal", "", "The calibration image")
var imgPath = flag.String("img", "", "The second image")
var dist	= flag.Int("dist", 0, "Distance from camera to trackpoint")
var width	= flag.Int("width", 0, "The diameter of the trackpoint in centimeters")

var savefile= flag.String("file", "", "File to save/load data to/from.")

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
	} else if *savefile != "" {
		load(&window, *savefile)
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
	if *calPath != "" {
		calimg,err := getImage(*calPath)
		if err == nil {
			window.Calibration = Frame{Path: *calPath, img: calimg}
		}
	}

	// Create X11 window
	xwindow,err := x11.NewWindow()
	window.Window = xwindow
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	window.Screen = xwindow.Screen()
	redraw(window, window.Frames[window.Cfra])

	memno := 1

	// Event loop
	for {
		event, ok := <-xwindow.EventChan()
		if ok != true {
			break
		}

		switch e := event.(type) {
		case gui.ConfigEvent:
			window.Screen = xwindow.Screen()
			redraw(window, window.Frames[window.Cfra])
			break
		case gui.MouseEvent:
			redraw(window, window.Frames[window.Cfra])
			break
		case gui.KeyEvent:
			fmt.Printf("Key: %d\n", e.Key)
			switch e.Key {
			case 'c':
			case 's':
				err := save(&window,*savefile)
				if err != nil {
					fmt.Printf("Error while saving: %s\n", err)
				}
			case 'l':
				err := load(&window, *savefile)
				if err != nil {
					fmt.Printf("Error while loading: %s\n", err)
				}
			case 'm':
				if *memprof == "" {
					break
				}
				f,_ := os.Create(fmt.Sprintf("%s%d",*memprof,memno))
				pprof.WriteHeapProfile(f)
				f.Close()
				memno++
			case 'a': // Find all circles
				go func() {
					fmt.Println("Trying to transform all frames .. ")
					for ; GetState(&window) == WORKING ; {
						time.Sleep(1e9)
					}

					frame := window.Cfra
					for i := 0 ; i < window.FrameCount ; i++ {
						if window.Frames[frame].Calculated {
							continue
						}
						window.Frames[frame].Centre = findCircle(&window,frame)
						window.Frames[frame].Calculated = true
						frame++
						runtime.GC()
					}
					fmt.Println("All frames transformed!")
				} ()
			case 'g':
				fmt.Println("Garbage collecting")
				runtime.GC()
			case KEY_SPACE: // space
				go func() {
					cfra := window.Cfra

					if window.Frames[cfra].Calculated {
						return
					}

					for ; GetState(&window) == WORKING ; {
						time.Sleep(1e9)
					}

					window.Frames[cfra].Centre = findCircle(&window,cfra)
					window.Frames[cfra].Calculated = true
					redraw(window, window.Frames[cfra])
					runtime.GC()
				} ()
			case KEY_LEFT:
				if window.Cfra > 0 {
					window.Cfra--
				}
				redraw(window, window.Frames[window.Cfra])
			case KEY_RIGHT:
				if window.Cfra < window.FrameCount-1 {
					window.Cfra++
				}
				redraw(window, window.Frames[window.Cfra])
			default:
			}
		}
	}
}
