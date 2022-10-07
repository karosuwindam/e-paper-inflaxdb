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

func ESetup() (*epaper.EPaper, error) {
	epd, err := epaper.New(M2in7bw)
	if err != nil {
		return nil, err
	}
	return epd, nil
}
func textPut(epd *epaper.EPaper, x, y int, text []string, size float64) {
	bufferReader := bytes.NewReader(writedata(text, size))
	image, _, err := image.Decode(bufferReader)
	if err != nil {
		// FIXME Better error handling.
		panic(err)
	}
	epd.AddLayer(image, x, y, true)
	epd.PrintDisplay()
}

func writedata(text []string, size float64) []byte {
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
		Size:    float64(fontsize),
		DPI:     0,
		Hinting: 0,

		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}
	imageWidth := 0
	for i := 0; i < len(text); i++ {
		imageWidth_t := 0.0

		slice := strings.Split(text[i], "")
		for _, str := range slice {
			if len(str) == 1 {
				imageWidth_t += fontsize*0.5 + fontsize*0.1
			} else {
				imageWidth_t += fontsize + fontsize*0.05
			}
		}
		if imageWidth < int(imageWidth_t) {
			imageWidth = int(imageWidth_t)

		}

	}
	imageHeight := int(fontsize * float64(len(text)))
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
	for i := 0; i < len(text); i++ {
		// dr.Dot.X = (fixed.I(imageWidth) - dr.MeasureString(text[i])) / 2
		dr.Dot.X = 0
		dr.Dot.Y = fixed.I(textTopMargin + int(fontsize)*i)
		dr.DrawString(text[i])
	}
	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return buf.Bytes()
}
