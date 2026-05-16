package main

import (
	"log"

	"github.com/chrispritchard/oled-clock/internal/sh1106"
)

func main() {
	disp, err := sh1106.NewSH1106()
	if err != nil {
		log.Fatalf("Failed to initialize SH1106: %v", err)
	}

	disp.Init()
	disp.Clear()

	var buffer [8][128]byte

	for x := 32; x < 96; x++ {
		buffer[3][x] = 0xff
		buffer[4][x] = 0xff
	}

	disp.ShowImage(buffer)

	select {}
}
