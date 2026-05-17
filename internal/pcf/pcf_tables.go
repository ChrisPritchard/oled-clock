package pcf

import (
	"encoding/binary"
	"fmt"
)

const (
	lenTOC                = 16
	lenCompressedMetric   = 5
	lenUncompressedMetric = 10

	minLenCompressedMetricTable   = 6
	minLenUncompressedMetricTable = 8

	minLenEncodingsTable = 14
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

type CompressedMetric struct {
	left_sided_bearing uint8
	right_side_bearing uint8
	character_width    uint8
	character_ascent   uint8
	character_descent  uint8
}

func parseCompressedMetric(source []byte) CompressedMetric {
	return CompressedMetric{
		left_sided_bearing: source[0],
		right_side_bearing: source[1],
		character_width:    source[2],
		character_ascent:   source[3],
		character_descent:  source[4],
	}
}

type UncompressedMetric struct {
	left_sided_bearing uint16
	right_side_bearing uint16
	character_width    uint16
	character_ascent   uint16
	character_descent  uint16
}

func parseUncompressedMetric(source []byte) UncompressedMetric {
	return UncompressedMetric{
		left_sided_bearing: binary.BigEndian.Uint16(source[0:2]),
		right_side_bearing: binary.BigEndian.Uint16(source[2:4]),
		character_width:    binary.BigEndian.Uint16(source[4:6]),
		character_ascent:   binary.BigEndian.Uint16(source[6:8]),
		character_descent:  binary.BigEndian.Uint16(source[8:10]),
	}
}

type CompressedMetricsTable struct {
	format        uint32
	metrics_count uint16
	metrics       []CompressedMetric
}

func parseCompressedMetricsTable(source []byte) (CompressedMetricsTable, error) {
	if len(source) < minLenCompressedMetricTable {
		return CompressedMetricsTable{}, fmt.Errorf("CompressedMetricTable parsing failure: expected at least %d bytes, but got %d", minLenCompressedMetricTable, len(source))
	}

	format := lsbint32(source[0:4])
	count := binary.BigEndian.Uint16(source[4:6])

	expectedLen := int(minLenCompressedMetricTable + (count * lenCompressedMetric))
	if len(source) != expectedLen {
		return CompressedMetricsTable{}, fmt.Errorf("CompressedMetricTable parsing failure: expected %d bytes, but got %d", expectedLen, len(source))
	}

	metrics := make([]CompressedMetric, count)

	for i := range count {
		os := minLenCompressedMetricTable + i*lenCompressedMetric
		metric := parseCompressedMetric(source[os : os+lenCompressedMetric])
		metrics[i] = metric
	}

	return CompressedMetricsTable{
		format:        format,
		metrics_count: count,
		metrics:       metrics,
	}, nil
}

type UncompressedMetricsTable struct {
	format        uint32
	metrics_count uint32
	metrics       []UncompressedMetric
}

func parseUncompressedMetricsTable(source []byte) (UncompressedMetricsTable, error) {
	if len(source) < minLenUncompressedMetricTable {
		return UncompressedMetricsTable{}, fmt.Errorf("UncompressedMetricTable parsing failure: expected at least %d bytes, but got %d", minLenCompressedMetricTable, len(source))
	}

	format := lsbint32(source[0:4])
	count := binary.BigEndian.Uint32(source[4:6])

	expectedLen := int(minLenUncompressedMetricTable + (count * lenUncompressedMetric))
	if len(source) != expectedLen {
		return UncompressedMetricsTable{}, fmt.Errorf("UncompressedMetricTable parsing failure: expected %d bytes, but got %d", expectedLen, len(source))
	}

	metrics := make([]UncompressedMetric, count)

	for i := range count {
		os := minLenUncompressedMetricTable + i*lenUncompressedMetric
		metric := parseUncompressedMetric(source[os : os+lenUncompressedMetric])
		metrics[i] = metric
	}

	return UncompressedMetricsTable{
		format:        format,
		metrics_count: count,
		metrics:       metrics,
	}, nil
}

type EncodingsTable struct {
	format            uint32
	min_char_or_byte2 uint16
	max_char_or_byte2 uint16
	min_byte1         uint16
	max_byte1         uint16
	default_char      uint16
	glyphindeces      []uint16
}

func parseEncodingsTable(source []byte) (EncodingsTable, error) {
	if len(source) < minLenEncodingsTable {
		return EncodingsTable{}, fmt.Errorf("EncodingsTable parsing failure: expected at least %d bytes, but got %d", minLenCompressedMetricTable, len(source))
	}

	format := lsbint32(source[0:4])
	min_char_or_byte2 := binary.BigEndian.Uint16(source[4:6])
	max_char_or_byte2 := binary.BigEndian.Uint16(source[6:8])
	min_byte1 := binary.BigEndian.Uint16(source[8:10])
	max_byte1 := binary.BigEndian.Uint16(source[10:12])
	default_char := binary.BigEndian.Uint16(source[12:14])

	count := (max_char_or_byte2 - min_char_or_byte2 + 1) * (max_byte1 - min_byte1 + 1)

	expectedLen := int(minLenEncodingsTable + (count * 2))
	if len(source) != expectedLen {
		return EncodingsTable{}, fmt.Errorf("EncodingsTable parsing failure: expected %d bytes, but got %d", expectedLen, len(source))
	}

	indices := make([]uint16, count)

	for i := range count {
		os := minLenEncodingsTable + i*2
		indices[i] = binary.BigEndian.Uint16(source[os : os+2])
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
