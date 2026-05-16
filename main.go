package main

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"log"
	"strconv"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed roboto-mono.ttf
var font_bytes []byte

func main() {
	font_face, err := freetype.ParseFont(font_bytes)
	if err != nil {
		log.Fatal("Could not parse font:", err)
	}

	width, height := 128, 64
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	font := getFont(font_face, 12, color.White, img)

	disp, err := NewSH1106()
	if err != nil {
		log.Fatalf("Failed to initialize SH1106: %v", err)
	}

	disp.Init()
	disp.Clear()

	i := 0
	for {
		background := color.Black
		draw.Draw(img, img.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)

		font.DrawString("counter: "+strconv.Itoa(i), freetype.Pt(20, 20))

		i++
		time.Sleep(1 * time.Second)

		pages := getBuffer(img)
		disp.ShowImage(pages)
	}
}

func getFont(font_face *truetype.Font, font_size float64, colour color.Color, target *image.RGBA) *freetype.Context {
	context := freetype.NewContext()
	context.SetDPI(72)
	context.SetFont(font_face)
	context.SetFontSize(font_size)
	context.SetClip(target.Bounds())
	context.SetDst(target)
	context.SetSrc(image.NewUniform(colour)) // White text
	context.SetHinting(font.HintingNone)
	return context
}

func getBuffer(img image.Image) [8][128]byte {
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
