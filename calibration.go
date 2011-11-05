package main

import "image"

type CalibrationData struct {
	Frame Frame
	Centre image.Point
	Distance int
	Width int
	Ppc int // Pixels per centimeter
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
