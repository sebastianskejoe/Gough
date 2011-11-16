package main

import (
	"image"
	"image/draw"
	"image/color"

	"strconv"
)

type CalibrationData struct {
	Image		draw.Image
	Frame		Frame
	Centre		image.Point
	Distance	int
	Width		int
	Ppc			int // Pixels per centimeter
	Color		color.Color
}

func (c *CalibrationData) Calibrate() error {
	// Make sure circle is found
	if c.Frame.Calculated != true {
		err := c.Frame.Calculate()
		if err != nil {
			return err
		}
	}

	c.Ppc = int(float32(c.Frame.Centre.Radius*2) / float32(c.Width))
	return nil
}

func (c *CalibrationData) Update(w *Window) error {
	c.Frame.Path = w.IR.EvalAsString("set calpath")
	img,err := getImage(c.Frame.Path)
	if err != nil {
		w.SetError("Couldn't load calibration image: ", err)
		return err
	}
	c.Frame.img = img
	c.Image = image.NewRGBA(img.Bounds())
	draw.Draw(c.Image, img.Bounds(), image.Black, image.Point{0,0}, draw.Src)

	ch := make(chan *image.Point)
	go filter(&img, ch)
	for p := range ch {
		c.Image.Set(p.X, p.Y, color.White)
	}

	w.IR.UploadImage("calimg", c.Image)
	w.IR.Eval(".cal.canvas configure -width ",c.Image.Bounds().Max.X," -height ",c.Image.Bounds().Max.Y)
	w.IR.Eval(`wm geometry .cal ""`)
	return nil
}

func CreateCalibrationWindow(w *Window) {
	w.IR.RegisterCommand("updateCalWin", func () { w.Calibration.Update(w) })
	w.IR.RegisterCommand("updateColor", func () { UpdateColor(w) })
	w.IR.UploadImage("calimg", w.Calibration.Image)
	w.IR.Eval(`
toplevel .cal
grid [ttk::label .cal.lcal -text "Path to calibration frame"] -column 0 -row 0 -sticky nwes
grid [ttk::button .cal.choosepath -text "Choose" -command {set calpath [tk_getOpenFile] ; updateCalWin}] -column 1 -row 0 -sticky nwes

grid [ttk::label .cal.ldist -text "Distance from lens to calibration item"] -column 0 -row 1 -sticky nwes
grid [ttk::spinbox .cal.distance -from 0 -to 900 -textvariable distance] -column 1 -row 1 -sticky news

grid [ttk::label .cal.lwidth -text "Width of calibration item in cm"] -column 0 -row 2 -sticky nwes
grid [ttk::spinbox .cal.width -from 0 -to 900 -textvariable width] -column 1 -row 2 -sticky news

set color "#FFFFFF"
grid [ttk::button .cal.choosecol -text "Choose color" -command {set color [tk_chooseColor -initialcolor [set color]] ; updateColor}] -column 0 -row 3 -sticky news

grid [canvas .cal.canvas] -columnspan 2 -column 0 -row 4 -sticky nwes
.cal.canvas create image 0 0 -anchor nw -image calimg

grid [ttk::button .cal.docal -text "Calibrate" -command {calibrate}] -column 0 -row 5 -columnspan 2 -sticky news
`)
}

func within(a, b, ts uint32) bool {
	low := b-ts
	if low > b {
		low = 0
	}
	up := b+ts
	if up < b {
		up = 2<<31-1
	}

	if a >= low && a <= up {
		return true
	}
	return false
}

func UpdateColor(w *Window) {
	cs := w.IR.EvalAsString("set color")
	if len(cs) < 7 {
		return
	}
	r,_ := strconv.Btoui64(cs[1:3],16)
	g,_ := strconv.Btoui64(cs[3:5],16)
	b,_ := strconv.Btoui64(cs[5:7],16)
	w.Calibration.Color = color.NRGBA{uint8(r),uint8(g),uint8(b),255}
	FilterFunc = func (c color.Color) bool {
		red,green,blue,_ := c.RGBA()
		r,g,b,_ := w.Calibration.Color.RGBA()
		if within(red,r,4000) && within(green,g,4000) && within(blue,b,4000) {
			return true
		}
		return false
	}
	w.Calibration.Update(w)
}
