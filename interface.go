package main

import (
	"image"
	"image/draw"
	"exp/gui"
)

func redraw(win *Window) {
	w := (*win)

	if w.Frames[w.Cfra].img == nil {
		w.Frames[w.Cfra].img = getImage(w.Frames[w.Cfra].Path)
	}
	img := w.Frames[w.Cfra].img
	draw.Draw(w.Screen, img.Bounds(), img, image.Point{0,0}, draw.Over)

	if w.State == WORKING {
		draw.Draw(w.Screen, image.Rectangle{image.Point{0,0}, image.Point{20,20}}, image.NewColorImage(image.RGBAColor{0,255,0,255}), image.Point{0,0}, draw.Src)
	}

	DrawCircle(w.Screen, w.Frames[w.Cfra].Centre.Centre, w.Frames[w.Cfra].Centre.Radius, image.RGBAColor{0,255,255,255})
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
