package main

import (
	"image"
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
