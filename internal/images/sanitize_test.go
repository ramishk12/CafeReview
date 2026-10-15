package images

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fixture builders ---

func jfifSegment() []byte {
	return []byte{0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00}
}

// exifSegmentWithOrientation builds an APP1 EXIF segment with an orientation and
// a privacy string ("GPX!") so a test can prove the string is removed.
func exifSegmentWithOrientation(orientation int) []byte {
	tiff := []byte{'I', 'I', 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00}
	tiff = append(tiff, 0x02, 0x00) // two entries
	// Entry 1: Orientation (0x0112), SHORT, value = orientation.
	e1 := []byte{0x12, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, byte(orientation), 0x00, 0x00, 0x00}
	tiff = append(tiff, e1...)
	// Entry 2: ImageDescription (0x010E), ASCII, count 5, value at offset 38.
	e2 := []byte{0x0E, 0x01, 0x02, 0x00, 0x05, 0x00, 0x00, 0x00, 0x26, 0x00, 0x00, 0x00}
	tiff = append(tiff, e2...)
	tiff = append(tiff, 0x00, 0x00, 0x00, 0x00) // no next IFD
	tiff = append(tiff, []byte("GPX!\x00")...)  // the privacy payload

	payload := append([]byte("Exif\x00\x00"), tiff...)
	return segment(0xE1, payload)
}

func segment(marker byte, payload []byte) []byte {
	segLen := len(payload) + 2
	out := []byte{0xFF, marker, byte(segLen >> 8), byte(segLen)}
	return append(out, payload...)
}

func jpegWithMetadata(orientation int) []byte {
	out := []byte{0xFF, 0xD8}                                                     // SOI
	out = append(out, jfifSegment()...)                                           // APP0 (keep)
	out = append(out, exifSegmentWithOrientation(orientation)...)                 // APP1 (strip, keep orientation)
	out = append(out, segment(0xFE, []byte("secret comment"))...)                 // COM (strip)
	out = append(out, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00) // SOS header
	out = append(out, 0x12, 0x34)                                                 // scan data
	out = append(out, 0xFF, 0xD9)                                                 // EOI
	return out
}

func pngChunk(typ string, data []byte) []byte {
	out := make([]byte, 0, len(data)+12)
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	out = append(out, length[:]...)
	out = append(out, typ...)
	out = append(out, data...)
	out = append(out, 0, 0, 0, 0) // CRC (not verified by the sanitizer)
	return out
}

func pngWithText() []byte {
	ihdr := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0}
	out := append([]byte{}, pngSignature...)
	out = append(out, pngChunk("IHDR", ihdr)...)
	out = append(out, pngChunk("tEXt", []byte("Author\x00Jane Doe"))...) // privacy (strip)
	out = append(out, pngChunk("IDAT", []byte{0x08, 0xD7})...)
	out = append(out, pngChunk("IEND", nil)...)
	return out
}

func webpChunk(fourCC string, data []byte) []byte {
	out := make([]byte, 0, len(data)+9)
	out = append(out, fourCC...)
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(len(data)))
	out = append(out, size[:]...)
	out = append(out, data...)
	if len(data)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

func webpWithMetadata() []byte {
	// VP8X flags byte: EXIF (0x08) and XMP (0x04) set.
	vp8x := []byte{0x0C, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	body := append([]byte("WEBP"), webpChunk("VP8X", vp8x)...)
	body = append(body, webpChunk("EXIF", []byte("GPSDATA\x00"))...) // privacy (strip)
	body = append(body, webpChunk("VP8 ", []byte{1, 2, 3, 4})...)    // image data (keep)

	out := append([]byte("RIFF"), make([]byte, 4)...)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(body)))
	return append(out, body...)
}

// --- tests ---

func TestSanitizeJPEG(t *testing.T) {
	in := jpegWithMetadata(6)
	out, err := Sanitize(in)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "GPX!", "the privacy value must be removed")
	assert.NotContains(t, string(out), "secret comment", "comments must be removed")
	assert.True(t, bytes.Contains(out, jfifSegment()), "the JFIF marker is kept")
	assert.True(t, bytes.Contains(out, exifOrientationSegment(6)), "a minimal orientation is kept")
	assert.True(t, bytes.HasSuffix(out, []byte{0xFF, 0xD9}), "the image ends at EOI")
}

func TestSanitizeJPEGDropsTrailingData(t *testing.T) {
	in := append(jpegWithMetadata(1), []byte("<script>alert(1)</script>")...)
	out, err := Sanitize(in)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "<script>")
	assert.True(t, bytes.HasSuffix(out, []byte{0xFF, 0xD9}))
}

func TestSanitizeJPEGWithoutOrientationAddsNoEXIF(t *testing.T) {
	out, err := Sanitize(jpegWithMetadata(1))
	require.NoError(t, err)
	assert.NotContains(t, string(out), "Exif", "orientation 1 needs no EXIF")
}

func TestSanitizePNG(t *testing.T) {
	out, err := Sanitize(pngWithText())
	require.NoError(t, err)

	assert.NotContains(t, string(out), "Jane Doe", "text metadata must be removed")
	assert.True(t, bytes.Contains(out, []byte("IHDR")), "IHDR is kept")
	assert.True(t, bytes.Contains(out, []byte("IDAT")), "image data is kept")
	assert.True(t, bytes.Contains(out, []byte("IEND")), "IEND is kept")
}

func TestSanitizeWebP(t *testing.T) {
	out, err := Sanitize(webpWithMetadata())
	require.NoError(t, err)

	assert.NotContains(t, string(out), "EXIF", "the EXIF chunk must be removed")
	assert.NotContains(t, string(out), "GPSDATA", "the EXIF payload must be removed")
	assert.True(t, bytes.Contains(out, []byte("VP8 ")), "image data is kept")

	// The RIFF size must match the new file length.
	riffSize := binary.LittleEndian.Uint32(out[4:8])
	assert.Equal(t, uint32(len(out)-8), riffSize)

	// The VP8X flags byte must have the EXIF and XMP bits cleared.
	vp8x := bytes.Index(out, []byte("VP8X"))
	require.Positive(t, vp8x)
	assert.Equal(t, byte(0), out[vp8x+8]&0x0C)
}

func TestSanitizeRejectsUnknownFormat(t *testing.T) {
	_, err := Sanitize([]byte("not an image"))
	assert.Error(t, err)
}

func TestSanitizeRejectsTruncatedContainer(t *testing.T) {
	_, err := Sanitize([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 8})
	assert.Error(t, err)
}
