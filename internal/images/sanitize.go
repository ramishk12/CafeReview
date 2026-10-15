// Package images strips privacy-related metadata from uploaded images while
// keeping everything needed to display them (colour profiles and EXIF
// orientation).
//
// It works on the container only: it walks the JPEG segments, PNG chunks, or
// WebP chunks and never decodes pixel data. That keeps memory use near zero, so
// a decompression bomb cannot exhaust the server, and no dimension cap is
// needed.
package images

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Sanitize returns a copy of a JPEG, PNG, or WebP image with privacy metadata
// removed. Colour profiles, the JFIF/Adobe markers, and the EXIF orientation
// tag are preserved. Data appended after the end of the image is dropped.
func Sanitize(data []byte) ([]byte, error) {
	switch {
	case isJPEG(data):
		return sanitizeJPEG(data)
	case isPNG(data):
		return sanitizePNG(data)
	case isWebP(data):
		return sanitizeWebP(data)
	}
	return nil, errors.New("images: unsupported or unrecognised format")
}

func isJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}

func isPNG(data []byte) bool {
	return len(data) >= len(pngSignature) && bytes.Equal(data[:len(pngSignature)], pngSignature)
}

func isWebP(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

// --- JPEG ---

const (
	jpegSOI   = 0xD8
	jpegEOI   = 0xD9
	jpegSOS   = 0xDA
	jpegAPP0  = 0xE0
	jpegAPP1  = 0xE1
	jpegAPP2  = 0xE2
	jpegAPP14 = 0xEE
	jpegCOM   = 0xFE
)

// sanitizeJPEG keeps the image data and the JFIF (APP0), ICC (APP2), and Adobe
// (APP14) markers. It drops the other application markers and comments, which
// carry EXIF, XMP, IPTC, and free-text metadata, and re-adds a minimal EXIF
// segment that contains only the orientation.
func sanitizeJPEG(data []byte) ([]byte, error) {
	i := 2
	orientation := 0
	var kept [][]byte

	for {
		if i >= len(data) {
			return nil, errors.New("images: malformed JPEG: no end of image")
		}
		if data[i] != 0xFF {
			return nil, errors.New("images: malformed JPEG: expected a marker")
		}
		markerStart := i
		// Skip any fill bytes before the marker code.
		for i < len(data) && data[i] == 0xFF {
			i++
		}
		if i >= len(data) {
			return nil, errors.New("images: malformed JPEG: truncated marker")
		}
		marker := data[i]
		codeEnd := i + 1

		switch {
		case marker == jpegEOI:
			out := assembleJPEG(kept, orientation)
			return append(out, 0xFF, jpegEOI), nil

		case marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7):
			// Standalone markers (TEM, RSTn) have no length.
			kept = append(kept, data[markerStart:codeEnd])
			i = codeEnd

		case marker == jpegSOS:
			// The scan header is a normal segment; the entropy-coded data follows
			// and runs until the EOI marker. Append the whole tail through EOI.
			if codeEnd+2 > len(data) {
				return nil, errors.New("images: malformed JPEG: truncated scan header")
			}
			segLen := int(binary.BigEndian.Uint16(data[codeEnd : codeEnd+2]))
			segEnd := codeEnd + segLen
			if segLen < 2 || segEnd > len(data) {
				return nil, errors.New("images: malformed JPEG: bad scan header")
			}
			eoi := findEOI(data, segEnd)
			if eoi < 0 {
				return nil, errors.New("images: malformed JPEG: no end of image")
			}
			kept = append(kept, data[markerStart:eoi+2])
			return assembleJPEG(kept, orientation), nil

		default:
			if codeEnd+2 > len(data) {
				return nil, errors.New("images: malformed JPEG: truncated segment length")
			}
			segLen := int(binary.BigEndian.Uint16(data[codeEnd : codeEnd+2]))
			segEnd := codeEnd + segLen
			if segLen < 2 || segEnd > len(data) {
				return nil, errors.New("images: malformed JPEG: bad segment length")
			}

			switch {
			case marker == jpegAPP1:
				// Drop EXIF/XMP, but remember the orientation.
				if o, ok := exifOrientation(data[codeEnd+2 : segEnd]); ok {
					orientation = o
				}
			case marker == jpegCOM:
				// Drop comments.
			case marker == jpegAPP0, marker == jpegAPP2, marker == jpegAPP14:
				kept = append(kept, data[markerStart:segEnd])
			case marker >= 0xE0 && marker <= 0xEF:
				// Other application markers (IPTC, Photoshop, JUMBF, ...): drop.
			default:
				kept = append(kept, data[markerStart:segEnd])
			}
			i = segEnd
		}
	}
}

func assembleJPEG(kept [][]byte, orientation int) []byte {
	out := []byte{0xFF, jpegSOI}
	if orientation >= 2 && orientation <= 8 {
		out = append(out, exifOrientationSegment(orientation)...)
	}
	for _, seg := range kept {
		out = append(out, seg...)
	}
	return out
}

// findEOI returns the index of the EOI marker at or after pos. In entropy-coded
// data a 0xFF byte is always followed by 0x00 (stuffing) or a RSTn marker, so the
// first 0xFFD9 is the true end of the image.
func findEOI(data []byte, pos int) int {
	for i := pos; i+1 < len(data); i++ {
		if data[i] == 0xFF && data[i+1] == jpegEOI {
			return i
		}
	}
	return -1
}

// exifOrientation reads the Orientation tag (0x0112) from an EXIF APP1 payload.
func exifOrientation(payload []byte) (int, bool) {
	if len(payload) < 14 || string(payload[:6]) != "Exif\x00\x00" {
		return 0, false
	}
	tiff := payload[6:]
	var order binary.ByteOrder
	switch string(tiff[0:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 0, false
	}
	if order.Uint16(tiff[2:4]) != 0x002A {
		return 0, false
	}
	ifd := int(order.Uint32(tiff[4:8]))
	if ifd+2 > len(tiff) {
		return 0, false
	}
	count := int(order.Uint16(tiff[ifd : ifd+2]))
	pos := ifd + 2
	for k := 0; k < count; k++ {
		if pos+12 > len(tiff) {
			return 0, false
		}
		entry := tiff[pos : pos+12]
		if order.Uint16(entry[0:2]) == 0x0112 {
			if order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) < 1 {
				return 0, false
			}
			v := int(order.Uint16(entry[8:10]))
			if v >= 1 && v <= 8 {
				return v, true
			}
			return 0, false
		}
		pos += 12
	}
	return 0, false
}

// exifOrientationSegment builds a minimal APP1 EXIF segment holding only the
// orientation tag, so the image still displays the right way up.
func exifOrientationSegment(orientation int) []byte {
	tiff := make([]byte, 0, 26)
	tiff = append(tiff, 'I', 'I', 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00) // header, IFD0 at offset 8
	tiff = append(tiff, 0x01, 0x00)                                   // one entry
	entry := make([]byte, 12)
	binary.LittleEndian.PutUint16(entry[0:2], 0x0112) // Orientation
	binary.LittleEndian.PutUint16(entry[2:4], 0x0003) // SHORT
	binary.LittleEndian.PutUint32(entry[4:8], 1)      // count
	binary.LittleEndian.PutUint16(entry[8:10], uint16(orientation))
	tiff = append(tiff, entry...)
	tiff = append(tiff, 0x00, 0x00, 0x00, 0x00) // no next IFD

	payload := append([]byte("Exif\x00\x00"), tiff...)
	segLen := len(payload) + 2
	seg := []byte{0xFF, jpegAPP1, byte(segLen >> 8), byte(segLen)}
	return append(seg, payload...)
}

// --- PNG ---

// sanitizePNG drops text and timestamp chunks and any embedded EXIF. Colour and
// rendering chunks are kept.
func sanitizePNG(data []byte) ([]byte, error) {
	out := append([]byte{}, pngSignature...)
	i := len(pngSignature)
	for i+8 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[i : i+4]))
		typ := string(data[i+4 : i+8])
		if length < 0 || i+12+length > len(data) {
			return nil, errors.New("images: malformed PNG: bad chunk length")
		}
		chunkEnd := i + 12 + length
		switch typ {
		case "tEXt", "zTXt", "iTXt", "eXIf", "tIME":
			// Privacy metadata: drop.
		default:
			out = append(out, data[i:chunkEnd]...)
		}
		i = chunkEnd
		if typ == "IEND" {
			// Anything after IEND is dropped.
			return out, nil
		}
	}
	return nil, errors.New("images: malformed PNG: no IEND chunk")
}

// --- WebP ---

// sanitizeWebP drops the EXIF and XMP chunks and clears their flags in the VP8X
// header. ICC colour and the image/animation chunks are kept.
func sanitizeWebP(data []byte) ([]byte, error) {
	body := data[12:]
	var kept [][]byte
	i := 0
	for i+8 <= len(body) {
		fourCC := string(body[i : i+4])
		size := int(binary.LittleEndian.Uint32(body[i+4 : i+8]))
		if size < 0 {
			return nil, errors.New("images: malformed WebP: bad chunk size")
		}
		end := i + 8 + size + (size & 1) // chunks are padded to an even size
		if end > len(body) {
			return nil, errors.New("images: malformed WebP: chunk exceeds file")
		}
		if fourCC != "EXIF" && fourCC != "XMP " {
			kept = append(kept, append([]byte{}, body[i:end]...))
		}
		i = end
	}

	var newBody []byte
	for _, raw := range kept {
		if string(raw[0:4]) == "VP8X" && len(raw) >= 9 {
			// Clear the EXIF (0x08) and XMP (0x04) bits in the flags byte.
			raw[8] &^= 0x0C
		}
		newBody = append(newBody, raw...)
	}

	out := make([]byte, 0, 12+len(newBody))
	out = append(out, 'R', 'I', 'F', 'F')
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(4+len(newBody))) // "WEBP" + chunks
	out = append(out, sizeBuf...)
	out = append(out, 'W', 'E', 'B', 'P')
	out = append(out, newBody...)
	return out, nil
}
