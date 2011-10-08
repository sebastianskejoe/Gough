package main

import (
	"image"
	"fmt"
	"exp/gui"
	"image/draw"
	"os"
	"image/png"
)

type Window struct {
	Bg image.Image
	Screen draw.Image
	Centre Circle
	Window gui.Window
	Ppc int // pixel per centimeter
	Dist int // distance from camera to track point
	Cfra int // current frame
	Frames []Frame
}

type Frame struct {
	Path string
	Id int
	Centre Circle
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
