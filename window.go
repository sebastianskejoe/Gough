package main

import (
	"exp/gui"
	"fmt"
	"github.com/nsf/gothic"
	"image"
	"image/color"
	"image/draw"
	"sync"

	"strconv"
	"strings"
)

type Window struct {
	Screen      draw.Image
	Window      gui.Window
	Ppc         int // pixel per centimeter
	Cfra        int // current frame
	State       int
	FrameCount  int
	Frames      []Frame
	Calibration CalibrationData
	IR          *gothic.Interpreter
}

func NewWindow(w, h int) (*Window, error) {
	ir := gothic.NewInterpreter("")
	window := &Window{
		IR:				ir,
		Screen:			image.NewNRGBA(image.Rect(0, 0, w, h)),
		Calibration:	CalibrationData{Image: image.NewNRGBA(image.Rect(0,0,w,h))},
	}

	return window, nil
}

func (w *Window) DrawFrame(frame int) error {
	var f Frame
	var err error

	if frame >= w.FrameCount {
		fmt.Println("Not in range")
		return nil
	}

	f = w.Frames[frame]

	if f.img == nil {
		f.img, err = getImage(f.Path)
		if err != nil {
			return err
		}
	}
	w.Screen = image.NewNRGBA(f.img.Bounds())
	draw.Draw(w.Screen, f.img.Bounds(), f.img, image.Point{0, 0}, draw.Src)
	DrawCircle(w.Screen, f.Centre.Centre, f.Centre.Radius, color.RGBA{0, 255, 255, 255})
	return nil
}

func (w *Window) CreateGUI() {
	w.IR.UploadImage("screen", w.Screen)

	w.IR.RegisterCommand("nextframe", func() { w.NextFrame() })
	w.IR.RegisterCommand("prevframe", func() { w.PrevFrame() })
	w.IR.RegisterCommand("loadframes", func() { w.LoadFrames(w.IR.EvalAsString("set path")) })
	w.IR.RegisterCommand("calibrate", func() { w.Calibrate() })
	w.IR.RegisterCommand("findcircle", func() { w.CalculateCurrent() })
	w.IR.RegisterCommand("findall", func() { w.CalculateAll() })
	w.IR.RegisterCommand("makecalwin", func() { CreateCalibrationWindow(w) })
	w.IR.RegisterCommand("loadfile", func() {
		err := load(w, w.IR.EvalAsString("set file"))
		if err != nil {
			w.SetError("Error while loading: ",err)
		}
		w.Update(w.Cfra)
	})
	w.IR.Eval(fmt.Sprintf(`
grid [ttk::button .prev -text "Previous" -command {prevframe}] -column 0 -row 0 -sticky news
grid [ttk::button .next -text "Next" -command {nextframe}] -column 1 -row 0 -sticky nwes

grid [ttk::entry .path -textvariable path] -column 0 -row 1 -sticky nwes
grid [ttk::button .addpaths -text "Add frames" -command {loadframes}] -column 1 -row 1


grid [canvas .canvas] -columnspan 2 -column 0 -row 5 -sticky news
.canvas create image 0 0 -anchor nw -image screen

grid [ttk::label .framecounter -textvariable framecounter] -column 0 -row 6 -sticky news
grid [ttk::progressbar .progress -maximum 1] -column 1 -row 6 -sticky nwes

grid [ttk::button .loadfile -text "Load" -command {set file [tk_getOpenFile] ; loadfile}] -column 0 -row 7

bind . <Left> { prevframe }
bind . <Right> { nextframe }

bind .path <Return> { loadframes }

menu .menu

.menu add cascade -menu .menu.options -label Options
.menu add cascade -menu .menu.track -label Track

menu .menu.track
.menu.track add command -label "Find circle" -command findcircle
.menu.track add command -label "Find all circles" -command findall

menu .menu.options
.menu.options add command -label Calibrate -command makecalwin

. configure -menu .menu
	`))
}

func (w *Window) Update(frame int) {
	err := w.DrawFrame(frame)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IR.UploadImage("screen", w.Screen)
	w.IR.Eval(".canvas configure -width ", w.Screen.Bounds().Max.X, " -height ", w.Screen.Bounds().Max.Y)
	w.IR.Eval(`wm geometry . ""`)

	// Frame counter
	if w.GetState() == IDLE {
		w.IR.Eval(`set framecounter "`, w.Cfra+1, "/", w.FrameCount, `"`)
	} else {
		w.IR.Eval(`set framecounter "WORKING `, w.Cfra+1, "/", w.FrameCount, `"`)
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
	if w.Cfra+1 == w.FrameCount {
		return
	}
	w.Cfra++
	w.Update(w.Cfra)
}

func (w *Window) PrevFrame() {
	if w.Cfra-1 < 0 {
		return
	}
	w.Cfra--
	w.Update(w.Cfra)
}

func (w *Window) CalculateCurrent() {
	cfra := w.Cfra
	go func() {
		// Only calculate if we are not already calculating
		if w.GetState() != IDLE {
			w.SetError("Already working!")
			return
		}
		w.SetState(WORKING)
		defer w.SetState(IDLE)

		// Calculate circle
		err := w.Frames[cfra].Calculate()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Update window to show changes
		w.Update(cfra)
	}()
}

func (w *Window) CalculateAll() {
	w.IR.Eval(".progress configure -maximum ", w.FrameCount)
	go func() {
		// Only calculate if we are not already calculating
		if w.GetState() != IDLE {
			w.SetError("Already working!")
			return
		}
		w.SetState(WORKING)
		defer w.SetState(IDLE)

		// Loop thorugh each frame
		for frame := 0; frame < w.FrameCount; frame++ {
			err := w.Frames[frame].Calculate()
			if err != nil {
				w.SetError("Couldn't find circle of frame ",frame,"(",err,")")
				continue
			}
			w.IR.Eval(".progress configure -value ", frame+1)
		}
		w.IR.Eval(".progress configure -value ", 0)

		// Update to show changes
		w.Update(w.Cfra)
	}()
}

func (w *Window) Calibrate() {
	w.Calibration.Distance = w.IR.EvalAsInt("set distance")
	w.Calibration.Width = w.IR.EvalAsInt("set width")
	go func() {
		// Don't calculate if we are working
		if w.GetState() != IDLE {
			w.SetError("Already working!")
			return
		}
		w.SetState(WORKING)
		defer w.SetState(IDLE)

		err := w.Calibration.Calibrate()
		if err != nil {
			w.SetError("Error occured in calibration: ",err)
		}
	}()
}

func (w *Window) LoadFrames(imgPath string) {
	r := strings.TrimFunc(imgPath, TrimFunc)
	split := strings.Split(r[1:len(r)-1], "-")
	start, _ := strconv.Atoi(split[0])
	end, _ := strconv.Atoi(split[1])
	f := 0
	for i := start; i <= end; i++ {
		path := strings.Replace(imgPath, r, strconv.Itoa(i), -1)
		w.Frames = append(w.Frames, Frame{Id: f, Path: path})
		f++
	}
	w.FrameCount = f
	w.Update(w.Cfra)
}

func (w *Window) SetError(args ...interface{}) {
	w.IR.Eval(fmt.Sprintf(`tk_messageBox -message "%s" -icon error`,fmt.Sprint(args...)))
}
