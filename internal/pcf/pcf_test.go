package pcf

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed ProggyCleanSZ.pcf
var proggy []byte

func TestParsing(t *testing.T) {
	pcf, err := NewPCF(proggy)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(pcf)
}
