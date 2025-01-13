package epaperv2

import (
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"time"

	"github.com/anthonynsimon/bild/paint"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// Epd is basic struc for Waveshare eps2in13bc
type Epd struct {
	Width             int
	Height            int
	lineWidth         int
	StartTransmission byte
	port              spi.PortCloser
	spiConn           spi.Conn
	rstPin            gpio.PinIO
	dcPin             gpio.PinIO
	csPin             gpio.PinIO
	busyPin           gpio.PinIO

	Display draw.Image
}

// CreateEpd is constructor for Epd
func CreateEpd() Epd {
	e := Epd{
		// Width:  122,
		// Height: 250,
		Width:             176,
		Height:            264,
		StartTransmission: REG_WRITE_RAM_BW,
	}

	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.lineWidth = lineWidth

	//init Display cash
	e.Display = paint.FloodFill(
		image.NewRGBA(image.Rect(0, 0, e.Width, e.Height)),
		image.Point{0, 0}, color.RGBA64{255, 255, 255, 255}, 255)
	return e
}

func (e *Epd) Open() error {
	var err error
	_, err = host.Init()
	if err != nil {
		slog.Error("host Init error", "error", err)
		return err
	}
	// SPI
	e.port, err = spireg.Open("")
	if err != nil {
		slog.Error("SPI Open Error", "error", err)
		return err
	}

	e.spiConn, err = e.port.Connect(3000000*physic.Hertz, 0b00, 8)
	if err != nil {
		slog.Error("spi poort connect error", "error", err)
		return err
	}
	slog.Debug("SPI Open", "conn", e.spiConn)

	// GPIO - read
	if checkGpioMemo0() {
		e.rstPin = gpioreg.ByName(GPIO_RST_PI5)   // out
		e.dcPin = gpioreg.ByName(GPIO_DC_PI5)     // out
		e.csPin = gpioreg.ByName(GPIO_CS_PI5)     // out
		e.busyPin = gpioreg.ByName(GPIO_BUSY_PI5) // in
	} else {
		e.rstPin = gpioreg.ByName(GPIO_RST)   // out
		e.dcPin = gpioreg.ByName(GPIO_DC)     // out
		e.csPin = gpioreg.ByName(GPIO_CS)     // out
		e.busyPin = gpioreg.ByName(GPIO_BUSY) // in
	}
	return nil
}

// Close is closing pariph.io port
func (e *Epd) Close() {
	e.port.Close()
}

// reset epd
func (e *Epd) reset() {
	e.rstPin.Out(true)
	time.Sleep(200 * time.Millisecond)
	e.rstPin.Out(false)
	time.Sleep(5 * time.Millisecond)
	e.rstPin.Out(true)
	time.Sleep(200 * time.Millisecond)
}

// sendCommand sets DC ping low and sends byte over SPI
func (e *Epd) sendCommand(command byte) {
	e.dcPin.Out(false)
	e.csPin.Out(false)
	c := []byte{command}
	r := make([]byte, len(c))
	e.spiConn.Tx(c, r)
	e.csPin.Out(true)
	// e.readBusy()
}

// sendData sets DC ping high and sends byte over SPI
func (e *Epd) sendData(data byte) {
	e.dcPin.Out(true)
	e.csPin.Out(false)
	c := []byte{data}
	r := make([]byte, len(c))
	e.spiConn.Tx(c, r)
	e.csPin.Out(true)
	// e.readBusy()
}

// ReadBusy waits for epd
func (e *Epd) readBusy() {
	//
	// 1: idle
	// 0: busy
	for e.busyPin.Read() == gpio.High {
		time.Sleep(100 * time.Millisecond)
	}
}

// Sleep powers off the epd
func (e *Epd) Sleep() {
	e.executeCommandAndLog(REG_DEEP_SLEEP_MODE, "DEEP_SLEEP", []byte{0x01})
	time.Sleep(100 * time.Millisecond)
}

func (e *Epd) PrintDisplay(isHorizon bool) {
	imgArray := e.convert(isHorizon)

	e.sendCommand(e.StartTransmission)

	for _, b := range imgArray {
		e.sendData(b)
	}
	e.TurnDisplayOn()
}

// Display sends an image to epd
func (e *Epd) DisplayView(image []byte) {
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.sendCommand(0x24)
	for j := 0; j < e.Height; j++ {
		for i := 0; i < lineWidth; i++ {
			e.sendData(image[i+j*lineWidth])
		}
	}
	e.TurnDisplayOn()
}

var lutData4Gray = []byte{
	0x40, 0x48, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x8, 0x48, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x2, 0x48, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x20, 0x48, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xA, 0x19, 0x0, 0x3, 0x8, 0x0, 0x0,
	0x14, 0x1, 0x0, 0x14, 0x1, 0x0, 0x3,
	0xA, 0x3, 0x0, 0x8, 0x19, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x1C,
}

// Init starts the epd
func (e *Epd) Init() {
	//EPD hardware init start
	e.reset()
	e.readBusy()

	//SWRESET
	e.executeCommandAndLog(REG_SW_RESET, "SOFT_RESET", nil)
	e.readBusy()

	// set Ram-Y address start/end position s=0 e=264
	e.executeCommandAndLog(REG_SET_RAM_Y_SE, "SET_Y-RAM_START_END_POSITION", []byte{0x00, 0x00, 0x07, 0x01})

	// set RAM y address count to 0
	e.executeCommandAndLog(REG_SET_RAM_Y_ADDRESS_COUNTER, "SET Y-RAM COUNT TO", []byte{0x00, 0x00})

	//data entry mode
	e.executeCommandAndLog(REG_DATA_ENTRY_MODE_SETTING, "DATA_ENTRY_MODE", []byte{0x03})

	slog.Debug("INIT DONE")
	time.Sleep(100 * time.Millisecond)
}

func (e *Epd) executeCommandAndLog(command byte, log string, data []byte) {
	slog.Debug(log, "command", command, "data", data)
	e.sendCommand(command)
	for i := 0; i < len(data); i++ {
		e.sendData(data[i])
	}
}

// Clear sets epd display to white (0xFF)
func (e *Epd) Clear() {
	e.Display = paint.FloodFill(
		image.Rect(0, 0, e.Display.Bounds().Dx(), e.Display.Bounds().Dy()),
		image.Point{0, 0}, color.RGBA{255, 255, 255, 255}, 255)
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.sendCommand(e.StartTransmission)
	for i := 0; i < e.Height; i++ {
		for j := 0; j < lineWidth; j++ {
			e.sendData(0xFF)
		}
	}
	e.TurnDisplayOn()
}

func (e *Epd) CrearDisplayData() {
	e.Display = paint.FloodFill(
		image.Rect(0, 0, e.Display.Bounds().Dx(), e.Display.Bounds().Dy()),
		image.Point{0, 0}, color.RGBA{255, 255, 255, 255}, 255)

}

func (e *Epd) Black() {
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.sendCommand(e.StartTransmission)
	for i := 0; i < e.Height; i++ {
		for j := 0; j < lineWidth; j++ {
			e.sendData(0x00)
		}
	}
	e.TurnDisplayOn()

}

// TurnDisplayOn turn the epd on
func (e *Epd) TurnDisplayOn() {
	// e.sendCommand(REG_DISPLAY_UPDATE_CTL_2)
	// e.sendData(0xF7)
	e.executeCommandAndLog(
		REG_DISPLAY_UPDATE_CTL_2,
		"reg dispay update control 2",
		[]byte{0xF7},
	)
	e.sendCommand(REG_MASTER_ACTIVATION)
	e.readBusy()
}

// TurnDisplayOff turn the display off
func (e *Epd) TurnDisplayOff() {
	e.sendCommand(REG_DISPLAY_UPDATE_CTL_2)
	e.sendData(0xC7)
	e.sendCommand(REG_MASTER_ACTIVATION)
}

// フォルダの存在確認
func checkGpioMemo0() bool {
	if _, err := os.Stat(CHECKPASS); os.IsNotExist(err) {
		return false
	}
	return true
}
