package main

import (
	"image"
	"image/draw"
	"exp/gui"
)

func redraw(win *Window) {
	w := (*win)
	draw.Draw(w.Screen, w.Bg.Bounds(), w.Bg, image.Point{0,0}, draw.Over)
	DrawCircle(w.Screen, w.Centre.Centre, w.Centre.Radius, image.RGBAColor{0,255,255,255})
	w.Window.FlushImage()
}

func events(c <-chan interface{}, w *Window) {
	for {
		event, ok := <-c
		if ok != true {
			break
		}

		switch e := event.(type) {
		case gui.ConfigEvent:
			redraw(w)
			break
		case gui.MouseEvent:
			redraw(w)
			break
		}
	}
}
