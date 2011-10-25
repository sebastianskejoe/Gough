package main

import (
	"image"
	"image/color"
	"image/draw"
	"exp/gui"
	"os"
	"sync"
)

type Window struct {
	Screen draw.Image
	Window gui.Window
	Ppc int // pixel per centimeter
	Dist int // distance from camera to track point
	Cfra int // current frame
	State int
	FrameCount int
	Frames []Frame
	Calibration Frame
}

func (w *Window) DrawFrame(frame int) {
	var f Frame
	var err os.Error

	if frame >= w.FrameCount {
		return
	}

	f = w.Frames[frame]

	if f.img == nil {
		f.img,err = getImage(f.Path)
		if err != nil {
			return
		}
	}
	draw.Draw(w.Screen, f.img.Bounds(), f.img, image.Point{0,0}, draw.Over)

	if w.State == WORKING {
		draw.Draw(w.Screen, image.Rectangle{image.Point{0,0}, image.Point{20,20}},
					image.NewUniform(color.RGBA{0,255,0,255}), image.Point{0,0}, draw.Src)
	}

	DrawCircle(w.Screen, f.Centre.Centre, f.Centre.Radius, color.RGBA{0,255,255,255})
	w.Window.FlushImage()
}

const (
	IDLE = iota
	WORKING
)

var statemutex sync.Mutex

func (w *Window) SetState(state int) {
	statemutex.Lock()
	w.State = state
	statemutex.Unlock()
}

func (w *Window) GetState() int {
	statemutex.Lock()
	defer statemutex.Unlock()
	return w.State
}
