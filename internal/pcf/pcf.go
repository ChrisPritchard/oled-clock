package pcf

import (
	"encoding/binary"
	"fmt"
	"log"
)

func lsbint32(source []byte) uint32 {
	if len(source) != 4 {
		log.Fatalf("expected four bytes, got %d", len(source))
	}
	return binary.LittleEndian.Uint32(source)
}

func Parse(data []byte) (any, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("file too small")
	}

	// Check magic header
	if string(data[0:4]) != "\x01fcp" {
		return nil, fmt.Errorf("invalid PCF magic header")
	}

	tableCount := lsbint32(data[4:8])
	tocTables := make([]TOC, tableCount)
	for i := range tocTables {
		os := i * lenTOC
		toc, err := parseTOC(data[os+8 : os+8+lenTOC])
		if err != nil {
			return nil, err
		}
		tocTables[i] = toc
	}

	fmt.Println(tocTables)

	return nil, nil
}
