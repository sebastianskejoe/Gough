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

	"runtime"
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
	ir := gothic.NewInterpreter("")
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
	w.Screen = image.NewNRGBA(f.img.Bounds())
	draw.Draw(w.Screen, f.img.Bounds(), f.img, image.Point{0,0}, draw.Src)
	DrawCircle(w.Screen, f.Centre.Centre, f.Centre.Radius, color.RGBA{0,255,255,255})
	return nil
}

func (w *Window) CreateGUI() {
	w.IR.UploadImage("screen", w.Screen)

	w.IR.RegisterCommand("nextframe", func () {w.NextFrame()})
	w.IR.RegisterCommand("prevframe", func () {w.PrevFrame()})
	w.IR.RegisterCommand("loadframes", func () {w.LoadFrames(w.IR.EvalAsString("set path"))})
	w.IR.RegisterCommand("calibrate", func () {w.Calibrate(w.IR.EvalAsString("set calpath"))})
	w.IR.RegisterCommand("findcircle", func () {w.CalculateCurrent()})
	w.IR.RegisterCommand("findall", func () {w.CalculateAll()})
	w.IR.Eval(fmt.Sprintf(`
ttk::entry .filename
grid [ttk::button .prev -text "Previous" -command {prevframe}] -column 0 -row 0 -sticky news
grid [ttk::button .next -text "Next" -command {nextframe}] -column 1 -row 0 -sticky nwes

grid [ttk::entry .path -textvariable path] -column 0 -row 1 -sticky nwes
grid [ttk::button .addpaths -text "Add frames" -command {loadframes}] -column 1 -row 1

grid [ttk::entry .calibrate -textvariable calpath] -column 0 -row 2 -sticky nwes
grid [ttk::spinbox .distance -from 0 -to 900 -textvariable distance] -column 1 -row 2 -sticky news
grid [ttk::spinbox .width -from 0 -to 900 -textvariable width] -column 0 -row 3 -sticky news
grid [ttk::button .docalibrate -text "Calibrate" -command {calibrate}] -column 1 -row 3

grid [ttk::button .findcircle -text "Find circle" -command {findcircle}] -column 0 -row 4
grid [ttk::button .findall -text "Find all circles" -command {findall}] -column 1 -row 4

grid [canvas .canvas] -columnspan 2 -column 0 -row 5 -sticky news
.canvas create image 0 0 -anchor nw -image screen

grid [ttk::label .framecounter -textvariable framecounter] -column 0 -row 6 -sticky news
grid [ttk::progressbar .progress -maximum 1] -column 1 -row 6 -sticky nwes

bind . <Left> { prevframe }
bind . <Right> { nextframe }
	`))
}

func (w *Window) Update(frame int) {
	fmt.Println("### In update")
	err := w.DrawFrame(frame)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IR.UploadImage("screen", w.Screen)
	w.IR.Eval(".canvas configure -width ",w.Screen.Bounds().Max.X," -height ",w.Screen.Bounds().Max.Y)
	w.IR.Eval(`wm geometry [winfo parent .canvas] ""`)

	// Frame counter
	if w.GetState() == IDLE {
		w.IR.Eval(`set framecounter "`,w.Cfra+1,"/",w.FrameCount,`"`)
	} else {
		w.IR.Eval(`set framecounter "WORKING `,w.Cfra+1,"/",w.FrameCount,`"`)
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
		if w.Frames[cfra].Calculated {
			return
		}
		if w.GetState() != IDLE {
			w.IR.Eval(`tk_messageBox -message "Already working!" -icon error`)
			return
		}
		w.SetState(WORKING)
		w.Update(w.Cfra)
		c,err := findCircle(w.Frames[cfra].Path)
		w.SetState(IDLE)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Frames[cfra].Centre = c
		w.Frames[cfra].Calculated = true
		w.Update(cfra)
		runtime.GC()
	} ()
}

func (w *Window) CalculateAll() {
	w.IR.Eval(".progress configure -maximum ",w.FrameCount)
	go func () {
		if w.GetState() != IDLE {
			w.IR.Eval(`tk_messageBox -message "Already working!" -icon error`)
			return
		}
		w.SetState(WORKING)

		// Loop thorugh each frame
		for frame := 0 ; frame < w.FrameCount ; frame++ {
			if w.Frames[frame].Calculated {
				continue
			}
			c,err := findCircle(w.Frames[frame].Path)
			if err != nil {
				w.IR.Eval(fmt.Sprintf(`tk_messageBox -message "Couldn't find circle of frame %d - %s" -icon error`, frame, err))
				continue
			}
			// New cicle successfully found
			w.Frames[frame].Centre = c
			w.Frames[frame].Calculated = true
			runtime.GC()
			w.IR.Eval(".progress configure -value ", frame+1)
		}
		w.IR.Eval(".progress configure -value ", 0)

		w.SetState(IDLE)
		fmt.Println("Now we want to update")
		w.Update(w.Cfra)
	} ()
}

func (w *Window) Calibrate(path string) {
	w.Calibration.Path = path
	w.Dist = w.IR.EvalAsInt("set distance")
	width := w.IR.EvalAsInt("set width")
	go func () {
		if w.Calibration.Calculated {
			return
		}
		if w.GetState() != IDLE {
			w.IR.Eval(`tk_messageBox -message "Already working!" -icon error`)
			return
		}
		w.SetState(WORKING)
		w.Update(w.Cfra)
		c,err := findCircle(path)
		w.SetState(IDLE)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Calibration.Centre = c
		w.Calibration.Calculated = true
		w.Ppc = int(float32(w.Calibration.Centre.Radius*2)/float32(width))
		fmt.Println(w.Calibration.Centre.Radius, w.Ppc)
		w.Update(w.Cfra)
		runtime.GC()
	} ()
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
