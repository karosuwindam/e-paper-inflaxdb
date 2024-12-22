package epaperv2

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"testing"
	"time"
)

func TestEPaper(t *testing.T) {
	e := CreateEpd()
	if err := e.Open(); err != nil {
		t.Fatalf("open error %v", err.Error())
	}
	defer e.Close()
	defer e.Clear()
	e.Init()
	e.Clear()
	time.Sleep(3 * time.Second)

	fmt.Printf("Display\n")
	// e.Display(getData())
	e.Black()
	fmt.Printf("sleeping\n")
	time.Sleep(3 * time.Second)
	ctx := contextWriteWriteData(context.Background(), []string{"test", "test2"}, 20)
	bufferReader := bytes.NewReader(writedata(ctx))

	img, _, err := image.Decode(bufferReader)
	if err != nil {
		return
	}
	e.AddLayer(img, 0, 0, true)
	e.PrintDisplay(true)
	fmt.Printf("sleeping\n")
	time.Sleep(3 * time.Second)
	e.Clear()
	//re Open Test
	e.Close()
	time.Sleep(3 * time.Second)
	e.Open()
	ctx = contextWriteWriteData(context.Background(), []string{"test2", "test3"}, 20)
	bufferReader = bytes.NewReader(writedata(ctx))

	img, _, err = image.Decode(bufferReader)
	if err != nil {
		return
	}
	e.AddLayer(img, 0, 0, true)
	e.PrintDisplay(true)
	fmt.Printf("sleeping\n")
	time.Sleep(3 * time.Second)

}
