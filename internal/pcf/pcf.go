package pcf

import (
	"fmt"
	"image"
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

		switch toc.tocType {

		case PCF_METRICS:
			isCompressed := (toc.format & PCF_COMPRESSED_METRICS) != 0
			metrics, err := parseMetricsTable(data[toc.offset:toc.offset+toc.size], isCompressed)
			if err != nil {
				return PCF{}, err
			}
			pcf.metrics = metrics

		case PCF_BITMAPS:
			bitmaps, err := parseBitmapTable(data[toc.offset : toc.offset+toc.size])
			if err != nil {
				return PCF{}, err
			}
			pcf.bitmaps = bitmaps

		case PCF_BDF_ENCODINGS:
			encodings, err := parseEncodingsTable(data[toc.offset : toc.offset+toc.size])
			if err != nil {
				return PCF{}, err
			}
			pcf.encodings = encodings
		}
	}

	return pcf, nil
}

func (pcf *PCF) glyphData(r rune) ([]byte, Metric, error) {
	index := int(r)
	if len(pcf.encodings.glyphindeces) < index {
		return nil, Metric{}, fmt.Errorf("index outside of encodings")
	}

	metric := pcf.metrics.Metrics[index]

	i := pcf.encodings.glyphindeces[index]
	start := pcf.bitmaps.offsets[i]
	var end int
	if i == len(pcf.bitmaps.offsets)-1 {
		end = len(pcf.bitmaps.offsets)
	} else {
		end = pcf.bitmaps.offsets[i+1]
	}

	return pcf.bitmaps.bitmap_data[start:end], metric, nil
}

func (pcf *PCF) calculateStride(metrics Metric) int {
	storage_unit := 1 << ((pcf.bitmaps.format >> 4) & 3) // 1=bytes, 2=shorts, 4=ints
	scanline_pad := 1 << (pcf.bitmaps.format & 3)        // 1=bytes, 2=shorts, 4=ints

	bytes_per_row := (metrics.CharacterWidth + storage_unit*8 - 1) / (storage_unit * 8) * storage_unit

	// Round up to scanline pad boundary
	return ((bytes_per_row + scanline_pad - 1) / scanline_pad) * scanline_pad
}

func (pcf *PCF) GetString(s string, spacing int) (*image.Gray, error) {
	if s == "" {
		return &image.Gray{}, nil
	}

	runes := []rune(s)

	// First pass: calculate total dimensions and store glyph data
	type glyphInfo struct {
		data   []byte
		width  int
		height int
		stride int
	}

	glyphs := make([]glyphInfo, len(runes))
	totalWidth := 0
	maxHeight := 0

	for i, r := range runes {
		data, metric, err := pcf.glyphData(r)
		if err != nil {
			return nil, err
		}

		width := metric.CharacterWidth
		height := metric.CharacterAscent + metric.CharacterDescent
		stride := pcf.calculateStride(metric)

		glyphs[i] = glyphInfo{
			data:   data,
			width:  width,
			height: height,
			stride: stride,
		}

		totalWidth += width
		if i < len(runes)-1 {
			totalWidth += spacing
		}
		if height > maxHeight {
			maxHeight = height
		}
	}

	// Create final image
	finalImg := image.NewGray(image.Rect(0, 0, totalWidth, maxHeight))
	finalPixels := finalImg.Pix
	finalStride := finalImg.Stride

	// Second pass: render directly to final image
	currentX := 0
	for i, glyph := range glyphs {
		// Render this glyph
		for y := 0; y < glyph.height; y++ {
			for s := 0; s < glyph.stride; s++ {
				bits := glyph.data[y*glyph.stride+s]

				for k := 7; k >= 0; k-- {
					x := s*8 + (7 - k)
					if x >= glyph.width {
						break
					}

					if bits&(1<<byte(k)) != 0 {
						finalPixels[y*finalStride+currentX+x] = 255
					} else {
						finalPixels[y*finalStride+currentX+x] = 0
					}
				}
			}
		}

		// Move to next glyph position
		currentX += glyph.width
		if i < len(runes)-1 {
			currentX += spacing
		}
	}

	return finalImg, nil
}
