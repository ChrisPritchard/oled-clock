package sh1106

import (
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

const (
	WIDTH   = 128
	HEIGHT  = 64
	DC_PIN  = "24"
	RST_PIN = "25"
)

type SH1106 struct {
	width, height int
	spi           spi.Conn
	dc            gpio.PinIO
	rst           gpio.PinIO
}

func NewSH1106() (*SH1106, error) {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	spiPort, err := spireg.Open("")
	if err != nil {
		return nil, err
	}

	spiConn, err := spiPort.Connect(40*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}

	dc := gpioreg.ByName(DC_PIN)
	if dc == nil {
		log.Fatalf("Failed to find %s pin", DC_PIN)
	}
	rst := gpioreg.ByName(RST_PIN)
	if rst == nil {
		log.Fatalf("Failed to find %s pin", RST_PIN)
	}

	dc.Out(gpio.Low)
	rst.Out(gpio.Low)

	dev := &SH1106{
		spi: spiConn,
		dc:  dc,
		rst: rst,
	}

	return dev, nil
}

func pause() {
	time.Sleep(100 * time.Millisecond)
}

func (d *SH1106) reset() {
	d.rst.Out(gpio.High)
	pause()
	d.rst.Out(gpio.Low)
	pause()
	d.rst.Out(gpio.High)
	pause()
}

func (d *SH1106) command(command byte) {
	d.dc.Out(gpio.Low)
	d.spi.Tx([]byte{command}, nil)
}

func (d *SH1106) Init() {
	d.reset()
	// Initialize display
	d.reset()
	d.command(0xAE) // turn off oled panel
	d.command(0x02) // -set low column address
	d.command(0x10) // -set high column address
	d.command(0x40) // set start line address  Set Mapping RAM Display Start Line (0x00~0x3F)
	d.command(0x81) // set contrast control register
	d.command(0xA0) // Set SEG/Column Mapping
	d.command(0xC0) // Set COM/Row Scan Direction
	d.command(0xA6) // set normal display
	d.command(0xA8) // set multiplex ratio(1 to 64)
	d.command(0x3F) // 1/64 duty
	d.command(0xD3) // set display offset    Shift Mapping RAM Counter (0x00~0x3F)
	d.command(0x00) // not offset
	d.command(0xd5) // set display clock divide ratio/oscillator frequency
	d.command(0x80) // set divide ratio, Set Clock as 100 Frames/Sec
	d.command(0xD9) // set pre-charge period
	d.command(0xF1) // Set Pre-Charge as 15 Clocks & Discharge as 1 Clock
	d.command(0xDA) // set com pins hardware configuration
	d.command(0x12)
	d.command(0xDB) // set vcomh
	d.command(0x40) // Set VCOM Deselect Level
	d.command(0x20) // Set Page Addressing Mode (0x00/0x01/0x02)
	d.command(0x02) //
	d.command(0xA4) //  Disable Entire Display On (0xa4/0xa5)
	d.command(0xA6) //  Disable Inverse Display On (0xa6/a7)
	pause()
	d.command(0xAF) // turn on oled panel
}

func (d *SH1106) ShowImage(pages [8][128]byte) {
	for p := range 8 {
		d.command(0xB0 + byte(p)) // set page address
		d.command(0x02)           // set low column address
		d.command(0x10)           // set high column address
		d.dc.Out(gpio.High)

		for x := range 128 {
			d.spi.Tx([]byte{pages[p][x]}, nil)
		}
	}
}

func (d *SH1106) Clear() {
	var pages [8][128]byte
	for p := range 8 {
		for x := range 128 {
			pages[p][x] = 0x00
		}
	}
	d.ShowImage(pages)
}
