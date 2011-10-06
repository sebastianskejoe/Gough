package main

import (
	"fmt"
	"image"
	"image/png"
	"image/draw"
	"os"
	"time"
	"flag"
	"runtime/pprof"
	"exp/gui"
	"exp/gui/x11"
	"./hough"
	"./edges"
	"./geometry"
)

var bg image.Image
var screen draw.Image
var centre hough.Circle
var window gui.Window

func filter(img *image.Image, filtered chan<- *image.Point) {
	bx := (*img).Bounds().Max.X
	by := (*img).Bounds().Max.Y
	sent := 0
	for x := 0 ; x < bx ; x++ {
		for y := 0 ; y < by ; y++ {
			r,_,_,_ := (*img).At(x,y).RGBA()
			if r > 200 {
				filtered <- &image.Point{x,y}
				sent++
			}
		}
	}
	fmt.Printf("Sent %d times in filter()\n", sent)
	close(filtered)
}

func findBounds(img *image.Image) image.Rectangle {
	bx := (*img).Bounds().Max.X
	by := (*img).Bounds().Max.Y
	maxx,maxy,minx,miny := 0,0,bx,by

	for x := 0 ; x < bx ; x++ {
		for y := 0 ; y < by ; y++ {
			r,_,_,_ := (*img).At(x,y).RGBA()
			if r > 200 {
				if x > maxx { maxx = x }
				if x < minx { minx = x }
				if y > maxy { maxy = y }
				if y < miny { miny = y }
			}
		}
	}

	return image.Rect(minx,miny,maxx,maxy)
}

/*func collector(centres, centres2 <-chan hough.Circle, done chan bool) {
	bx, by := bg.Bounds().Max.X,bg.Bounds().Max.Y
	votes := make([]int, bx*by*bx)
	var centre,c hough.Circle
	max := 0


L:
	for {
		select {
		case i1,ok := <-centres:
			if ok != true {
				break L
			}
			c = i1
		case i2,ok := <-centres2:
			if ok != true {
				break L
			}
			c = i2
		}
		index := c.Centre.X*bx+c.Centre.Y*by+c.Radius
		votes[index] += 1
		if votes[index] > max {
			max = votes[index]
			centre = c
		}
//		fmt.Printf("%d %d %d\n", centre.coord.x, centre.coord.y, centre.radius)
	}
	fmt.Printf("(%d,%d) - %d\n", centre.Centre.X, centre.Centre.Y, centre.Radius)
//	end := time.Seconds()
//	fmt.Printf("Done in %d\n", end-start)
	done <- true
}*/

func redraw() {
	draw.Draw(screen, bg.Bounds(), bg, image.Point{0,0}, draw.Over)
	geometry.Circle(screen, centre.Centre, centre.Radius, image.RGBAColor{0,255,255,255})
	window.FlushImage()
}

func events(c <-chan interface{}) {
	for {
		event, ok := <-c
		if ok != true {
			break
		}

		switch e := event.(type) {
		case gui.ConfigEvent:
			redraw()
			break
		case gui.MouseEvent:
			redraw()
			break
		}
	}
}

var cpuprof = flag.String("cpuprofile", "", "Cpu profile")
var drawGui = flag.Bool("gui", true, "Make a GUI")

func main() {
	file,err := os.Open("fig1.png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer file.Close()

	img,err := png.Decode(file)
	bg = img
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	flag.Parse()
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			goto L
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
L:

	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *hough.Circle, 400)

	// Create X11 window
	window,err = x11.NewWindow()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	screen = window.Screen()
	redraw()
	go events(window.EventChan())

	start := time.Seconds()

	r := findBounds(&img)

	mid := time.Seconds()

	go filter(&img,filtered)
	go edges.Sobel(&img,filtered,edge)
	go hough.Transform(&img, edge, transformed, r)

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
			centre = *c
		}
	}

	geometry.Circle(screen, centre.Centre, centre.Radius, image.RGBAColor{0,255,255,255})
	window.FlushImage()

	end := time.Seconds()
	fmt.Printf("Done in %d seconds, findBounds() took %d seconds\n", end-start, mid-start)

	for {
	}
}
