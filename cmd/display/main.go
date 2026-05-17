package main

import (
	_ "embed"
	"image"
	"image/draw"
	"log"
	"strings"
	"time"

	"github.com/chrispritchard/oled-clock/internal/pcf"
	"github.com/chrispritchard/oled-clock/internal/sh1106"
)

//go:embed ProggyCleanSZ.pcf
var proggy []byte

func main() {
	font, err := pcf.NewPCF(proggy)
	if err != nil {
		log.Fatal(err)
	}

	disp, err := sh1106.NewSH1106()
	if err != nil {
		log.Fatalf("Failed to initialize SH1106: %v", err)
	}

	disp.Init()
	disp.Clear()

	nzt := timeZone("Pacific/Auckland")
	hkt := timeZone("Asia/Hong_Kong")

	tick := false
	for {
		now := time.Now()

		dst := image.NewGray(image.Rect(0, 0, 128, 64))
		offset := image.Pt(10, 10)

		nzt_img, err := timeImage(now, nzt, tick, font)
		if err != nil {
			log.Fatal(err)
		}
		draw.Draw(dst, nzt_img.Rect.Add(offset), nzt_img, image.Point{}, draw.Src)

		hkt_img, err := timeImage(now, hkt, tick, font)
		if err != nil {
			log.Fatal(err)
		}
		draw.Draw(dst, hkt_img.Rect.Add(offset).Add(image.Pt(0, 14)), hkt_img, image.Point{}, draw.Src)

		date_img, err := font.GetString(now.Format("02 January 2006"), 0)
		draw.Draw(dst, date_img.Rect.Add(offset).Add(image.Pt(0, 30)), date_img, image.Point{}, draw.Src)

		disp.ShowImage(dst)
		time.Sleep(1 * time.Second)
		tick = !tick
	}
}

func timeImage(now time.Time, zone *time.Location, tick bool, font *pcf.PCF) (*image.Gray, error) {
	rel := now.In(zone)
	name, _ := rel.Zone()
	if len(name) == 3 {
		name += " "
	}
	time_string := rel.Format("03:04 PM")
	if tick {
		time_string = strings.Replace(time_string, ":", " ", 1)
	}
	time_string = name + " " + time_string

	return font.GetString(time_string, 0)
}

func timeZone(iana string) *time.Location {
	loc, err := time.LoadLocation(iana)
	if err != nil {
		log.Fatal(err)
	}
	return loc
}
