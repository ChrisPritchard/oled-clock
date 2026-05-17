package pcf

import (
	_ "embed"
	"image/png"
	"os"
	"testing"
)

//go:embed ProggyCleanSZ.pcf
var proggy []byte

func TestParsing(t *testing.T) {
	pcf, err := NewPCF(proggy)
	if err != nil {
		t.Error(err)
	}

	img, err := pcf.GetString("Hello World!", 20)
	if err != nil {
		t.Error(err)
	}

	// ignored in git
	f, err := os.Create("test_output.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
