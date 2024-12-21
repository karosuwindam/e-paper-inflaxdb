package epaperv2

import (
	"bytes"
	"fmt"
	"image"
	"testing"
	"time"
)

func TestEPaper(t *testing.T) {
	e := CreateEpd()
	defer e.Close()
	defer e.Clear()
	e.Init()
	e.Clear()
	time.Sleep(5 * time.Second)

	fmt.Printf("Display\n")
	// e.Display(getData())
	e.Black()
	fmt.Printf("sleeping\n")
	time.Sleep(5 * time.Second)
	bufferReader := bytes.NewReader(writedata(strWriteData{[]string{"test"}, 20}))

	image, _, err := image.Decode(bufferReader)
	if err != nil {
		return
	}

	e.AddLayer(image, 0, 0, true)

	e.PrintDisplay(true)
	fmt.Printf("sleeping\n")
	time.Sleep(5 * time.Second)

}
