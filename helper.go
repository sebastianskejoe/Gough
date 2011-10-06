package main

import (
	"image"
	"fmt"
	"exp/gui"
	"image/draw"
)

type Window struct {
	Bg image.Image
	Screen draw.Image
	Centre Circle
	Window gui.Window
}

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
