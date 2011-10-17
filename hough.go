package main

import (
	"image"
	"math"
	"fmt"
)


func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/* Performs the hough transformation
* pixels: pixels to transform
* centres: calculated circles
* b: all pixels sent along pixels are within b, meaning that the centre must be inside of b. This assumes high quality data.*/
func Transform(pixels <-chan image.Point, centres chan<- *Circle, b image.Rectangle) {
	max := max(b.Max.X-b.Min.X, b.Max.Y-b.Min.Y)
	// Radius cannot be bigger than max/, so a+b cannot be bigger than max
	maxr := float64(max/2*max/2)
	sent := 0

	for pixel := range pixels {
		px,py := pixel.X, pixel.Y

		for x := b.Min.X ; x < b.Max.X ; x++ {
			for y := b.Min.Y ; y < b.Max.Y ; y++ {

				// Much much faster than doing math.Pow(..., 2)
				a := float64((px-x)*(px-x))
				b := float64((py-y)*(py-y))
				if a+b > maxr {
					continue
				}

				r := int(math.Sqrt((a+b)))
				if r < 0 || r > max {
					continue
				}
				c := new(Circle)
				c.Centre = image.Point{x,y}
				c.Radius = r
				centres <- c
				sent++
			}
		}
	}
	fmt.Printf("Sent %d times in hough.Transform()\n", sent)
	close(centres)
}
