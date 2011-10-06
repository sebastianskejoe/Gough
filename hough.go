package main

import (
	"image"
	"math"
	"fmt"
)

type Circle struct {
	Centre image.Point
	Radius int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/* Performs the hough transformation
 * Edge pixels to transform */
func Transform(img *image.Image, pixels <-chan image.Point, centres chan<- *Circle, b image.Rectangle) {
//	imgx,imgy := (*img).Bounds().Max.X, (*img).Bounds().Max.Y
	max := max(b.Max.X-b.Min.X, b.Max.Y-b.Min.Y)
	// Radius cannot be bigger than max/, so a+b cannot be bigger than max
	maxr := float64(max/2*max/2)
	sent := 0

	for {
		pixel,ok := <-pixels
		if ok == false {
			break
		}
		px,py := pixel.X, pixel.Y

		for x := b.Min.X ; x < b.Max.X ; x++ {
			for y := b.Min.Y ; y < b.Max.X ; y++ {

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
				centres <- &Circle{Centre: image.Point{x,y}, Radius: r}
				sent++
			}
		}
	}
	fmt.Printf("Sent %d times in hough.Transform()\n", sent)
	close(centres)
}
