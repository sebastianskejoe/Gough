package main

import (
	"image"
	"fmt"
	"os"
	"image/png"
	"image/color"
	"time"
)

func filter(img *image.Image, filtered chan<- *image.Point) {
	bx := (*img).Bounds().Max.X
	by := (*img).Bounds().Max.Y
//	sent := 0
	for x := 0 ; x < bx ; x++ {
		for y := 0 ; y < by ; y++ {
			if ColorIsGood((*img).At(x,y)) {
				filtered <- &image.Point{x,y}
//				sent++
			}
		}
	}
//	fmt.Printf("Sent %d times in filter()\n", sent)
	close(filtered)
}

func ColorIsGood(c color.Color) bool {
	r,_,_,_ := c.RGBA()
	if r > 30000 {
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

func getImage(path string) (image.Image,os.Error) {
	// Get image
	file,err := os.Open(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil,err
	}
	defer file.Close()

	img,err := png.Decode(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil,err
	}
	return img,err
}

func TrimFunc(c int) bool {
	if c == '[' || c == ']' {
		return false
	}
	return true
}

func findCircle(window *Window, frame int) Circle {
	filtered := make(chan *image.Point, 1000)
	edge := make(chan image.Point, 100)
	transformed := make(chan *Circle, 400)

	img,_ := getImage(window.Frames[frame].Path)

	fmt.Printf("Finding circle of %s ... ", window.Frames[frame].Path)

	window.SetState(WORKING)
	window.DrawFrame(frame)

	start := time.Seconds()

	r := findBounds(&img)

	go filter(&img,filtered)
	go Sobel(&img,filtered,edge)
	go Transform(edge, transformed, r)

	bx,by := img.Bounds().Max.X,img.Bounds().Max.Y
	votes := make([]int, bx*by*bx)
	votes[0] = 1
	var c *Circle
	var centre Circle
	var index int
	max := 0

	for c = range transformed {
		index = c.Radius+bx*c.Centre.Y+bx*by*c.Centre.X
		votes[index] += 1
		if votes[index] > max {
			max = votes[index]
			centre = *c
		}
	}

	end := time.Seconds()
	fmt.Printf("Done in %d seconds\n", end-start)

	window.SetState(IDLE)

	return centre
}
