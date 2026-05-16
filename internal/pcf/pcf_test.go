package pcf

import (
	_ "embed"
	"testing"
)

//go:embed ProggyCleanSZ.pcf
var proggy []byte

func TestParsing(t *testing.T) {
	Parse(proggy)
}
