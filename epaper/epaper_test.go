package epaper

import "testing"

func TestEPaper(t *testing.T) {
	if err := initEpaper(); err != nil {
		return
	}
	device.ClearScreen()

}
