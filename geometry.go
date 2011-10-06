package geometry

import (
	"image/draw"
	"image"
)

func absf(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}


/* Line drawing - Bresenhams algorithm in it's simplest form.*/
func Line(screen draw.Image, a image.Point, b image.Point, color image.Color) {
	dx := b.X-a.X
	dy := b.Y-a.Y
	error := 0.0
	derr := absf(float64(dy)/float64(dx))
	y := a.Y
	for x := a.X ; x < b.X ; x++ {
		screen.Set(x, y, color)
		error += derr
		if error >= 0.5 {
			y++
			error -= 1.0
		}
	}
}


func Circle(screen draw.Image, c image.Point, r int, color image.Color) {
	f := 1-r
	ddF_x := 1
	ddF_y := -2*r
	x,y := 0,r

	screen.Set(c.X, c.Y+r, color)
	screen.Set(c.X, c.Y-r, color)
	screen.Set(c.X+r, c.Y, color)
	screen.Set(c.X-r, c.Y, color)

	for ; x < y ; {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x

		screen.Set(c.X+x, c.Y+y, color)
		screen.Set(c.X-x, c.Y+y, color)

		screen.Set(c.X+x, c.Y-y, color)
		screen.Set(c.X-x, c.Y-y, color)

		screen.Set(c.X+y, c.Y+x, color)
		screen.Set(c.X-y, c.Y+x, color)

		screen.Set(c.X+y, c.Y-x, color)
		screen.Set(c.X-y, c.Y-x, color)
	}
}


