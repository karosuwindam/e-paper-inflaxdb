package epaper

import (
	"bytes"
	"context"
	"epaperifdb/config"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log/slog"
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

var device *epaper.EPaper = nil

func initEpaper() error {
	if device != nil {
		return nil
	}
	epd, err := epaper.New(M2in7bw)
	if err != nil {
		slog.Error("Failed to initialize e-paper device", "error", err.Error())
		return err
	}
	slog.Debug("E-paper device initialized")
	epd.Init()
	if config.OutURL.InitClear {
		epd.ClearScreen()
	}
	device = epd
	return nil
}

func testPut(ctx context.Context, x, y int, text []string, size float64) error {
	ctx, span := config.TracerS(ctx, "epaper.testPut", "epaper")
	defer span.End()
	slog.DebugContext(ctx, "testPut", "x", x, "y", y, "text", text, "size", size)

	ctx = contextWriteWriteData(ctx, text, size)
	bufferReader := bytes.NewReader(writedata(ctx))
	image, _, err := image.Decode(bufferReader)
	if err != nil {
		// FIXME Better error handling.
		return err
	}
	device.AddLayer(image, x, y, true)
	device.PrintDisplay()
	return nil
}

type strWriteData struct {
	text []string
	size float64
}

func writedata(ctx context.Context) []byte {
	ctx, span := config.TracerS(ctx, "epaper.writedata", "epaper")
	defer span.End()
	data, ok := contextReadWriteData(ctx)
	if !ok {
		slog.ErrorContext(ctx, "Failed to get data from context")
		return nil
	}
	// フォントファイルを読み込み
	ft, err := fontfileRead(TRUETYPE_GOTHIC)
	if err != nil {
		slog.ErrorContext(ctx, "FontFile read error",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	fontsize := data.size
	texts := data.text

	slog.DebugContext(ctx, "Font Parse success",
		"size", fontsize,
		"text", texts,
	)

	img := createImage(ft, fontsize, texts)
	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		slog.ErrorContext(ctx, "Image Encode to png err", "error", err.Error())
		os.Exit(1)
	}
	return buf.Bytes()
}

const (
	TRUETYPE_GOTHIC = "/usr/share/fonts/truetype/fonts-japanese-gothic.ttf"
)

func fontfileRead(fontfilepass string) (*truetype.Font, error) {
	ftBinary, err := ioutil.ReadFile(fontfilepass)
	if err != nil {
		return nil, err
	}
	return truetype.Parse(ftBinary)
}

func setImageWidth(fontsize float64, texts []string) int {
	imageWidth := 0
	for i := 0; i < len(texts); i++ {
		imageWidth_t := 0.0

		slice := strings.Split(texts[i], "")
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
	return imageWidth
}

func createImage(ft *truetype.Font, fontsize float64, texts []string) *image.RGBA {

	opt := truetype.Options{
		Size:    fontsize,
		DPI:     0,
		Hinting: 0,

		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}

	imageWidth := setImageWidth(fontsize, texts)
	imageHeight := int(fontsize * float64(len(texts)))
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
	for i := 0; i < len(texts); i++ {
		dr.Dot.X = 0
		dr.Dot.Y = fixed.I(textTopMargin + int(fontsize)*i)
		dr.DrawString(texts[i])
	}
	return img
}
