package main

import (
	"fmt"
	"flag"
)

var savefile = flag.String("file", "", "File to save/load data to/from.")

func main() {
	window,err := NewWindow(1,1)
	if err != nil {
		fmt.Println(err)
		return
	}
	flag.Parse()

	// Parse image paths
	if *savefile != "" {
		load(window, *savefile)
	}

	window.CreateGUI()
	window.Update(0)
	<-window.IR.Done
}
