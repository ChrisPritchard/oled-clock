package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/chrispritchard/pcf"
)

//go:embed ProggyCleanSZ.pcf
var proggy []byte

func main() {

	pf, e := pcf.Parse(proggy)
	if e != nil {
		log.Fatal(e)
	}
	b, _, _, _ := pf.Lookup('h')
	fmt.Println(b)

	// disp, err := sh1106.NewSH1106()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize SH1106: %v", err)
	// }

	// disp.Init()
	// disp.Clear()

	// var buffer [8][128]byte

	// disp.ShowImage(buffer)

	// select {}
}
