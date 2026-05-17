package pcf

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	lenTOC                = 16
	lenCompressedMetric   = 5
	lenUncompressedMetric = 12

	minLenCompressedMetricTable   = 6
	minLenUncompressedMetricTable = 8

	minLenEncodingsTable = 14

	minLenBitmapTable = 22
)

func lsbint32(source []byte) uint32 {
	if len(source) != 4 {
		log.Fatalf("expected four bytes, got %d", len(source))
	}
	return binary.LittleEndian.Uint32(source)
}

func byteorder(format uint32) binary.ByteOrder {
	if (format & 4) == 0 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

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

type Metric struct {
	LeftSidedBearing    uint16
	RightSideBearing    uint16
	CharacterWidth      int
	CharacterAscent     int
	CharacterDescent    int
	CharacterAttributes uint16
}

func parseCompressedMetric(source []byte) Metric {
	return Metric{
		LeftSidedBearing: uint16(source[0] - 0x80),
		RightSideBearing: uint16(source[1] - 0x80),
		CharacterWidth:   int(source[2] - 0x80),
		CharacterAscent:  int(source[3] - 0x80),
		CharacterDescent: int(source[4] - 0x80),
	}
}

func parseUncompressedMetric(source []byte, byteOrder binary.ByteOrder) Metric {
	return Metric{
		LeftSidedBearing:    byteOrder.Uint16(source[0:2]),
		RightSideBearing:    byteOrder.Uint16(source[2:4]),
		CharacterWidth:      int(byteOrder.Uint16(source[4:6])),
		CharacterAscent:     int(byteOrder.Uint16(source[6:8])),
		CharacterDescent:    int(byteOrder.Uint16(source[8:10])),
		CharacterAttributes: byteOrder.Uint16(source[10:12]),
	}
}

type MetricsTable struct {
	Format       uint32
	MetricsCount int
	Metrics      []Metric
}

func parseMetricsTable(source []byte, isCompressed bool) (MetricsTable, error) {
	tableLen := minLenCompressedMetricTable
	metricLen := lenCompressedMetric

	if !isCompressed {
		tableLen = minLenUncompressedMetricTable
		metricLen = lenUncompressedMetric
	}

	if len(source) < tableLen {
		return MetricsTable{}, fmt.Errorf("MetricsTable parsing failure: expected at least %d bytes, got %d", tableLen, len(source))
	}

	format := lsbint32(source[0:4])
	byteOrder := byteorder(format)

	var metricsCount int

	if isCompressed {
		metricsCount = int(byteOrder.Uint16(source[4:6]))
	} else {
		metricsCount = int(byteOrder.Uint32(source[4:8]))
	}

	expectedLen := tableLen + metricsCount*metricLen
	if len(source) < expectedLen {
		return MetricsTable{}, fmt.Errorf("MetricsTable parsing failure: expected at least %d bytes for compressed metrics, got %d", expectedLen, len(source))
	}

	metrics := make([]Metric, metricsCount)
	for i := range metricsCount {
		offset := tableLen + int(i)*metricLen
		data := source[offset : offset+metricLen]
		if isCompressed {
			metrics[i] = parseCompressedMetric(data)
		} else {
			metrics[i] = parseUncompressedMetric(data, byteOrder)
		}
	}

	return MetricsTable{
		Format:       format,
		MetricsCount: metricsCount,
		Metrics:      metrics,
	}, nil
}

type EncodingsTable struct {
	format            uint32
	min_char_or_byte2 uint16
	max_char_or_byte2 uint16
	min_byte1         uint16
	max_byte1         uint16
	default_char      uint16
	glyphindeces      []int
}

func parseEncodingsTable(source []byte) (EncodingsTable, error) {
	if len(source) < minLenEncodingsTable {
		return EncodingsTable{}, fmt.Errorf("EncodingsTable parsing failure: expected at least %d bytes, but got %d", minLenEncodingsTable, len(source))
	}

	format := lsbint32(source[0:4])
	byteOrder := byteorder(format)

	min_char_or_byte2 := byteOrder.Uint16(source[4:6])
	max_char_or_byte2 := byteOrder.Uint16(source[6:8])
	min_byte1 := byteOrder.Uint16(source[8:10])
	max_byte1 := byteOrder.Uint16(source[10:12])
	default_char := byteOrder.Uint16(source[12:14])

	count := int((max_char_or_byte2 - min_char_or_byte2 + 1) * (max_byte1 - min_byte1 + 1))

	expectedLen := minLenEncodingsTable + (count * 2)
	if len(source) < expectedLen {
		return EncodingsTable{}, fmt.Errorf("EncodingsTable parsing failure: expected at least %d bytes, but got %d", expectedLen, len(source))
	}

	indices := make([]int, count)

	for i := range count {
		os := minLenEncodingsTable + int(i)*2
		indices[i] = int(byteOrder.Uint16(source[os : os+2]))
	}

	return EncodingsTable{
		format:            format,
		min_char_or_byte2: min_char_or_byte2,
		max_char_or_byte2: max_char_or_byte2,
		min_byte1:         min_byte1,
		max_byte1:         max_byte1,
		default_char:      default_char,
		glyphindeces:      indices,
	}, nil
}

type BitmapTable struct {
	format      uint32
	glyph_count uint32
	offsets     []int
	bitmapSizes [4]uint32
	bitmap_data []uint8
}

func parseBitmapTable(source []byte) (BitmapTable, error) {
	if len(source) < minLenBitmapTable {
		return BitmapTable{}, fmt.Errorf("BitmapTable parsing failure: expected at least %d bytes, but got %d", minLenBitmapTable, len(source))
	}

	format := lsbint32(source[0:4])
	byteOrder := byteorder(format)

	glyph_count := byteOrder.Uint32(source[4:8])

	if len(source) < minLenBitmapTable+int(glyph_count)*4 {
		return BitmapTable{}, fmt.Errorf("BitmapTable parsing failure: offsets array exceeds source length")
	}

	offsets := make([]int, glyph_count)
	for i := range glyph_count {
		os := 8 + i*4
		offsets[i] = int(byteOrder.Uint32(source[os : os+4]))
	}

	cur := 8 + int(glyph_count)*4

	sizes := [4]uint32{
		byteOrder.Uint32(source[cur : cur+4]),
		byteOrder.Uint32(source[cur+4 : cur+8]),
		byteOrder.Uint32(source[cur+8 : cur+12]),
		byteOrder.Uint32(source[cur+12 : cur+16]),
	}

	bitmapSizeIndex := format & 3
	if bitmapSizeIndex > 3 {
		return BitmapTable{}, fmt.Errorf("BitmapTable parsing failure: invalid bitmap size index %d", bitmapSizeIndex)
	}

	expectedSize := int(sizes[bitmapSizeIndex])
	bitmapDataStart := cur + 16

	if len(source) < bitmapDataStart+expectedSize {
		return BitmapTable{}, fmt.Errorf("BitmapTable parsing failure: bitmap data truncated. Expected %d bytes, have %d",
			bitmapDataStart+expectedSize, len(source))
	}

	bitmap_data := make([]uint8, expectedSize)
	copy(bitmap_data, source[bitmapDataStart:bitmapDataStart+int(expectedSize)])

	return BitmapTable{
		format:      format,
		glyph_count: glyph_count,
		offsets:     offsets,
		bitmapSizes: sizes,
		bitmap_data: bitmap_data,
	}, nil
}
