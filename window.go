package main

import (
	"image"
	"image/color"
	"image/draw"
	"exp/gui"
	"os"
	"sync"
	"github.com/nsf/gothic"
	"fmt"

	"strings"
	"strconv"
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
	IR *gothic.Interpreter
}

func NewWindow(w,h int) (*Window, os.Error) {
	ir,err := gothic.NewInterpreter()
	if err != nil {
		return nil, err
	}
	window := &Window{
		IR: ir,
		Screen: image.NewNRGBA(image.Rect(0,0,w,h)),
	}

	return window,nil
}

func (w *Window) DrawFrame(frame int) os.Error {
	var f Frame
	var err os.Error

	if frame >= w.FrameCount {
		fmt.Println("Not in range")
		return nil
	}

	f = w.Frames[frame]

	if f.img == nil {
		f.img,err = getImage(f.Path)
		if err != nil {
			return err
		}
	}
	draw.Draw(w.Screen, f.img.Bounds(), f.img, image.Point{0,0}, draw.Over)
	DrawCircle(w.Screen, f.Centre.Centre, f.Centre.Radius, color.RGBA{0,255,255,255})
	return nil
}

var fc *gothic.StringVar

func (w *Window) CreateGUI() {
	w.IR.UploadImage("screen", w.Screen)

	path := w.IR.NewStringVar("path")

	fc = w.IR.NewStringVar("framecounter")

	w.IR.RegisterCallback("nextframe", func () {w.NextFrame()})
	w.IR.RegisterCallback("prevframe", func () {w.PrevFrame()})
	w.IR.RegisterCallback("loadframes", func () {w.LoadFrames(path.Get())})
	w.IR.RegisterCallback("findcircle", func () {w.CalculateCurrent()})
	w.IR.Eval(fmt.Sprintf(`
entry .filename
grid [button .prev -text "Previous" -command {prevframe}] -column 0 -row 0 -sticky news
grid [button .next -text "Next" -command {nextframe}] -column 1 -row 0 -sticky nwes
grid [entry .path -textvariable path] -column 0 -row 1 -sticky nwes
grid [button .addpaths -text "Add frames" -command {loadframes}] -column 1 -row 1
grid [button .findcircle -text "Find circle" -command {findcircle}] -column 0 -row 2
grid [label .canvas -image screen] -columnspan 1 -column 0 -row 3 -sticky news

grid [label .framecounter -textvariable framecounter] -columnspan 2 -column 0 -row 4 -sticky news

bind . <Left> { prevframe }
bind . <Right> { nextframe }
	`))
}

func (w *Window) Update(frame int) {
	err := w.DrawFrame(frame)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IR.UploadImage("screen", w.Screen)

	// Frame counter
	if w.GetState() == IDLE {
		fc.Set(fmt.Sprintf("%d/%d",w.Cfra+1,w.FrameCount))
	} else {
		fc.Set(fmt.Sprintf("WORKING! %d/%d", w.Cfra+1, w.FrameCount))
	}
}


//// Window state

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

//// Simple functions to call from the interface
func (w *Window) NextFrame() {
	if w.Cfra+1 < w.FrameCount {
		w.Cfra++
	}
	w.Update(w.Cfra)
}

func (w *Window) PrevFrame() {
	if w.Cfra-1 >= 0 {
		w.Cfra--
	}
	w.Update(w.Cfra)
}

func (w *Window) CalculateCurrent() {
	cfra := w.Cfra
	go func() {
		if w.GetState() != IDLE {
			w.IR.AsyncEval(`tk_messageBox -message "Already working!" -icon error`)
			return
		}
		w.SetState(WORKING)
		w.IR.Async(func () {w.Update(w.Cfra)},nil,nil)
		c,err := findCircle(w, cfra)
		w.SetState(IDLE)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Frames[cfra].Centre = c
		w.Frames[cfra].Calculated = true
		w.IR.Async(func () {
			w.Update(cfra)
		}, nil, nil)
	} ()
}

func (w *Window) CalculateAll() {
}

func (w *Window) LoadFrames(imgPath string) {
	r := strings.TrimFunc(imgPath, TrimFunc)
	split := strings.Split(r[1:len(r)-1], "-")
	start,_ := strconv.Atoi(split[0])
	end,_ := strconv.Atoi(split[1])
	f := 0
	for i := start ; i <= end ; i++ {
		path := strings.Replace(imgPath, r, strconv.Itoa(i),-1)
		w.Frames = append(w.Frames, Frame{Id: f, Path: path})
		f++
	}
	w.FrameCount = f
	w.Update(w.Cfra)
}
