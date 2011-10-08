package main

import (
	"image"
	"fmt"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clampedFilter(c image.Color) int {
	if ColorIsGood(c) { return 255 }
	return 0
}

func Sobel(imgP *image.Image, pixels <-chan *image.Point, edge chan<- image.Point) {
	img := *imgP
	sent := 0
	for {
		pixel,ok := <-pixels
		if ok == false {
			break
		}
		x,y := pixel.X,pixel.Y

		gx := clampedFilter(img.At(x-1,y-1))+2*clampedFilter(img.At(x,y-1))+clampedFilter(img.At(x+1,y-1)) -
				clampedFilter(img.At(x-1,y+1))-2*clampedFilter(img.At(x,y+1))-clampedFilter(img.At(x+1,y+1))
		gy := -clampedFilter(img.At(x-1,y-1))+clampedFilter(img.At(x+1,y-1))-
				2*clampedFilter(img.At(x-1,y))+2*clampedFilter(img.At(x+1,y))-
				clampedFilter(img.At(x-1,y+1))+clampedFilter(img.At(x+1,y+1))

		g := abs(gx)+abs(gy)
		if g >= 255 {
			edge <-*pixel
			sent++
		}
	}
	fmt.Printf("Sent %d times in edges.Sobel\n", sent)
	close(edge)
}
