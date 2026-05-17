package pcf

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	PCF_METRICS       = 1 << 2
	PCF_BITMAPS       = 1 << 3
	PCF_BDF_ENCODINGS = 1 << 5

	PCF_DEFAULT            = 0x00000000
	PCF_COMPRESSED_METRICS = 0x00000100
)

type PCF struct {
	metrics   MetricsTable
	bitmaps   BitmapTable
	encodings EncodingsTable
}

func lsbint32(source []byte) uint32 {
	if len(source) != 4 {
		log.Fatalf("expected four bytes, got %d", len(source))
	}
	return binary.LittleEndian.Uint32(source)
}

func NewPCF(data []byte) (PCF, error) {
	if len(data) < 8 {
		return PCF{}, fmt.Errorf("file too small")
	}

	// magic header
	if string(data[0:4]) != "\x01fcp" {
		return PCF{}, fmt.Errorf("invalid PCF magic header")
	}

	var pcf PCF

	tableCount := lsbint32(data[4:8])
	for i := range tableCount {
		os := i * lenTOC
		toc, err := parseTOC(data[os+8 : os+8+lenTOC])
		if err != nil {
			return PCF{}, err
		}

		// only need three tables: PCF_METRICS (sizes), PCF_BITMAPS (pixel data) and PCF_BDF_ENCODINGS (mappings to characters)

		if toc.tocType == PCF_METRICS {
			isCompressed := (toc.format & PCF_COMPRESSED_METRICS) != 0
			metrics, err := parseMetricsTable(data[toc.offset:toc.offset+toc.size], isCompressed)
			if err != nil {
				return PCF{}, err
			}
			pcf.metrics = metrics
		}

		if toc.tocType == PCF_BITMAPS {
			bitmaps, err := parseBitmapTable(data[toc.offset : toc.offset+toc.size])
			if err != nil {
				return PCF{}, err
			}
			pcf.bitmaps = bitmaps
		}

		if toc.tocType == PCF_BDF_ENCODINGS {
			encodings, err := parseEncodingsTable(data[toc.offset : toc.offset+toc.size])
			if err != nil {
				return PCF{}, err
			}
			pcf.encodings = encodings
		}
	}

	return pcf, nil
}
