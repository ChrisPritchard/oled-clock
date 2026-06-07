package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"log"
	"net"
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
	i := 0
	cur_bat, e := get_bat()
	last_bat := -1
	if e == nil {
		last_bat = cur_bat
	}

	for {
		now := time.Now()

		dst := image.NewGray(image.Rect(0, 0, 128, 64))
		offset := image.Pt(10, 8)

		// NZ time
		nzt_img, err := timeImage(now, nzt, tick, font)
		if err != nil {
			log.Fatal(err)
		}
		draw.Draw(dst, nzt_img.Rect.Add(offset), nzt_img, image.Point{}, draw.Src)

		// HK time
		hkt_img, err := timeImage(now, hkt, tick, font)
		if err != nil {
			log.Fatal(err)
		}
		draw.Draw(dst, hkt_img.Rect.Add(offset).Add(image.Pt(0, 14)), hkt_img, image.Point{}, draw.Src)

		// date
		date_img, err := font.GetString(now.Format("2 Jan"), 0)
		if err != nil {
			log.Fatal(err)
		}
		draw.Draw(dst, date_img.Rect.Add(offset).Add(image.Pt(0, 30)), date_img, image.Point{}, draw.Src)

		// bat
		if last_bat != -1 && i%5 == 0 {
			i = 0
			last_bat, _ = get_bat()
		}
		if last_bat != -1 {
			bat_img, err := font.GetString(fmt.Sprintf("bat:%d%%", last_bat), 0)
			if err != nil {
				log.Fatal(err)
			}
			draw.Draw(dst, bat_img.Rect.Add(offset).Add(image.Pt(55, 35)), bat_img, image.Point{}, draw.Src)
		}

		disp.ShowImage(dst)
		tick = !tick
		i++
		time.Sleep(1 * time.Second)
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
	if time_string[0] == '0' {
		time_string = " " + time_string[1:]
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

func get_bat() (int, error) {
	conn, err := net.Dial("tcp", "127.0.0.1:8423")
	if err != nil {
		return 0, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("get battery\n"))
	if err != nil {
		return 0, fmt.Errorf("failed to send command: %w", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var level int
	_, err = fmt.Sscanf(response, "battery: %d", &level)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return level, nil
}
