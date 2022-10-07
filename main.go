package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/otaviokr/go-epaper-lib"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	M2in7bw = epaper.Model{Width: 176, Height: 264, StartTransmission: 0x13}
)

func main() {
	epd, err := Setup()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	epd.Init()
	epd.ClearScreen()
	textPut(epd, 0, 0, "こんにちは", 30)
}

func Setup() (*epaper.EPaper, error) {

	epd, err := epaper.New(M2in7bw)
	if err != nil {
		return nil, err
	}
	return epd, nil

}

func textPut(epd *epaper.EPaper, x, y int, text string, size float64) {
	bufferReader := bytes.NewReader(writedata(text, size))
	image, _, err := image.Decode(bufferReader)
	if err != nil {
		// FIXME Better error handling.
		panic(err)
	}
	epd.AddLayer(image, x, y, true)
	epd.PrintDisplay()

}

func writedata(text string, size float64) []byte {

	// フォントファイルを読み込み
	ftBinary, err := ioutil.ReadFile("/usr/share/fonts/truetype/fonts-japanese-gothic.ttf")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ft, err := truetype.Parse(ftBinary)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fontsize := size
	opt := truetype.Options{
		Size:              float64(fontsize),
		DPI:               0,
		Hinting:           0,
		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}

	slice := strings.Split(text, "")
	imageWidth_t := 0.0
	for _, str := range slice {
		if len(str) == 1 {
			imageWidth_t += fontsize*0.5 + fontsize*0.1
		} else {
			imageWidth_t += fontsize + fontsize*0.05
		}
	}
	imageWidth := int(imageWidth_t)
	imageHeight := int(fontsize)
	textTopMargin := int(fontsize * 0.9)

	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)

	face := truetype.NewFace(ft, &opt)

	dr := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{},
	}

	dr.Dot.X = (fixed.I(imageWidth) - dr.MeasureString(text)) / 2
	dr.Dot.Y = fixed.I(textTopMargin)

	dr.DrawString(text)

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return buf.Bytes()
}
