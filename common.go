package main

import (
	"exp/gui"
	"image"
	"image/draw"
)

const (
	IDLE = iota
	WORKING
)

const (
	KEY_SPACE = 32
	KEY_LEFT = 65361
	KEY_RIGHT = 65363
)

type Circle struct {
	Centre image.Point
	Radius int
}

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

type Frame struct {
	Path string
	Id int
	Centre Circle
	Calculated bool
	img image.Image
}
