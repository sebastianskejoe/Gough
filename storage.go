package main

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"
)

func load(w *Window, path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}

	r := bufio.NewReader(file)
	count := 0
	for line, err := r.ReadString('\n'); err == nil; line, err = r.ReadString('\n') {
		items := strings.Split(line, ":")
		id, _ := strconv.Atoi(items[0])
		fpath := items[1]
		x, _ := strconv.Atoi(items[2])
		y, _ := strconv.Atoi(items[3])
		radius, _ := strconv.Atoi(items[4][0 : len(items[4])-1])
		w.Frames = append(w.Frames, Frame{Id: id, Path: fpath, Centre: Circle{Centre: image.Point{x, y}, Radius: radius}})
		count++
	}
	w.FrameCount = count
	return nil
}

func save(w *Window, path string) error {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}

	for frame := 0; frame < w.FrameCount; frame++ {
		f := w.Frames[frame]
		fmt.Fprintf(file, "%d:%s:%d:%d:%d\n", f.Id, f.Path, f.Centre.Centre.X, f.Centre.Centre.Y, f.Centre.Radius)
	}

	return nil
}
