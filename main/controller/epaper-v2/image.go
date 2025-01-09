package epaperv2

import (
	"bytes"
	"epaperifdb/config"
	"image"
	"image/color"
	"image/draw"

	"github.com/anthonynsimon/bild/clone"
	"github.com/anthonynsimon/bild/transform"
	"github.com/icemint0828/imgedit"
)

// AddLayer puts img on top of the previous layers prepared to be printed. Function Clearscreen() will also delete any prepared layer.
func (e *Epd) AddLayer(img image.Image, startX, startY int, transparent bool) {
	selectionRectangle := image.Rect(startX, startY, startX+img.Bounds().Dx(), startY+img.Bounds().Dy())

	if transparent {
		mask := clone.AsRGBA(img)
		maskColor := color.RGBA{255, 255, 255, 255}
		for i := 0; i < mask.Bounds().Dx(); i++ {
			for j := 0; j < mask.Bounds().Dy(); j++ {
				if mask.RGBAAt(i, j) == maskColor {
					mask.SetRGBA(i, j, color.RGBA{255, 255, 255, 0})
				}
			}
		}
		draw.DrawMask(e.Display, selectionRectangle, image.NewUniform(color.RGBA{A: 255}), image.Point{X: 0, Y: 0}, mask, image.Point{X: 0, Y: 0}, draw.Over)
	} else {
		draw.Draw(e.Display, selectionRectangle, img, image.Point{X: 0, Y: 0}, draw.Src)
	}
}

// Convert the input image into a ready-to-display byte buffer.
func (e *Epd) convert(isHorizon bool) []byte {
	var dstImg image.Image
	var clearBackground byte = 0x00
	if config.OutURL.Mirror {
		c := imgedit.NewConverter(e.Display)
		c.Reverse(isHorizon)
		dstImg = c.Convert()
	} else {
		dstImg = e.Display
	}

	// Processing each line from the original image. If image is too large, we'll cap to the screen size.
	height := dstImg.Bounds().Dy()
	// height := e.Display.Bounds().Dy()
	// if e.display.Bounds().Dy() > e.model.Height {
	// 	height = e.model.Height
	// }
	width := dstImg.Bounds().Dx()
	// width := e.Display.Bounds().Dx()
	// if e.display.Bounds().Dx() > e.model.Width {
	// 	width = e.model.Width
	// }

	// Create the output array (each element represents 8 pixels, so we need a smaller array than the original matrix.)
	buffer := bytes.Repeat([]byte{0xFF}, e.lineWidth*e.Height)
	offset := 0
	var newValue byte = clearBackground
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			// Shift previous values before calculating the current one.
			newValue = newValue << 1

			// If color in pixel (x,y) is black, we mark it on the correct bit in the new element for the array.
			// if color.Palette([]color.Color{color.Black, color.White}).Index(e.Display.At(i, j)) == 1 {
			if color.Palette([]color.Color{color.Black, color.White}).Index(dstImg.At(i, j)) == 1 {
				newValue |= 0x01
			}

			// A new byte is ready to be appended to buffer array.
			if i > 0 && (i+1)%8 == 0 {
				buffer[(((i+1)/8)-1)+(j*width)+offset] = newValue
				newValue = clearBackground
			}
		}
		offset += e.lineWidth - width
	}

	return buffer
}

// Rotate will rotate the image 90 degrees clockwise. Use it before calling convert, because convert will insert the image in the display representation matrix.
func (e *Epd) Rotate(img image.Image) image.Image {
	return transform.Rotate(img, 90.0, &transform.RotationOptions{ResizeBounds: true, Pivot: &image.Point{0, 0}})
}
