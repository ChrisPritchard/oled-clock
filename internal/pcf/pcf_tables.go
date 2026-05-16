package pcf

import "fmt"

const (
	lenTOC = 16
)

type TOC struct {
	tocType uint32
	format  uint32
	size    uint32
	offset  uint32
}

func parseTOC(source []byte) (TOC, error) {
	if len(source) != lenTOC {
		return TOC{}, fmt.Errorf("TOC parsing failure: expected %d bytes, but got %d", lenTOC, len(source))
	}

	return TOC{
		tocType: lsbint32(source[0:4]),
		format:  lsbint32(source[4:8]),
		size:    lsbint32(source[8:12]),
		offset:  lsbint32(source[12:16]),
	}, nil
}
