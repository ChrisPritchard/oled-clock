package main

import (
	"image"
	"image/color"
	"log"
)

func main() {
	disp, err := NewSH1106()
	if err != nil {
		log.Fatalf("Failed to initialize SH1106: %v", err)
	}

	disp.Init()
	disp.Clear()

	square := drawSquare()
	pages := GetBuffer(square)
	disp.ShowImage(pages)

	select {}
}

func GetBuffer(img image.Image) [8][128]byte {
	var pages [8][128]byte

	for page := range 8 {
		startRow := page * 8

		for x := range 128 {
			var byteVal byte

			for bit := range 8 {
				y := startRow + bit
				if y >= 64 {
					break
				}

				r, g, b, _ := img.At(x, y).RGBA()
				if (r+g+b)/3 > 32768 {
					byteVal |= 1 << bit
				}
			}
			pages[page][x] = byteVal
		}
	}

	return pages
}

func drawSquare() *image.Gray {
	width, height := 128, 64
	img := image.NewGray(image.Rect(0, 0, width, height))

	squareSize := 40
	sx := (width - squareSize) / 2
	sy := (height - squareSize) / 2

	for y := range height {
		for x := range width {
			img.SetGray(x, y, color.Gray{0})
		}
	}

	for y := sy; y < sy+squareSize; y++ {
		for x := sx; x < sx+squareSize; x++ {
			img.SetGray(x, y, color.Gray{255})
		}
	}

	return img
}
