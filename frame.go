package main

import (
	"image"
	"runtime"
)

type Frame struct {
	Path string
	Id int
	Centre Circle
	Calculated bool
	img image.Image
}

func (f *Frame) Calculate() error {
	if f.Calculated {
		return nil
	}
	c,err := findCircle(f.Path)
	if err != nil {
		return err
	}
	// New cicle successfully found
	f.Centre = c
	f.Calculated = true
	runtime.GC()
	return nil
}
