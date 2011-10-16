package main

import (
	"image"
	"image/draw"
	"image/color"
	"os"
)

func redraw(w Window, f Frame) {

	var err os.Error
	if f.img == nil {
		f.img,err = getImage(f.Path)
		if err != nil {
			return
		}
	}
	draw.Draw(w.Screen, f.img.Bounds(), f.img, image.Point{0,0}, draw.Over)

	if w.State == WORKING {
		draw.Draw(w.Screen, image.Rectangle{image.Point{0,0}, image.Point{20,20}}, image.NewUniform(color.RGBA{0,255,0,255}), image.Point{0,0}, draw.Src)
	}

	DrawCircle(w.Screen, f.Centre.Centre, f.Centre.Radius, color.RGBA{0,255,255,255})
	w.Window.FlushImage()
}
