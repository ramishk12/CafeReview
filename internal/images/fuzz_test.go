package images

import (
	"bytes"
	"testing"
)

// FuzzSanitize throws arbitrary and mutated bytes at Sanitize. It asserts that
// the sanitizer never panics, keeps the format, and is idempotent.
//
// Run the full fuzzing phase on demand:
//
//	go test ./internal/images/ -run=FuzzSanitize -fuzz=FuzzSanitize -fuzztime=30s
//
// A normal `go test` only replays the seed corpus plus any regressions saved
// under testdata/fuzz.
func FuzzSanitize(f *testing.F) {
	f.Add(jpegWithMetadata(6))
	f.Add(jpegWithMetadata(1))
	f.Add(pngWithText())
	f.Add(webpWithMetadata())
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xD9}) // bare JPEG
	f.Add([]byte("not an image"))

	f.Fuzz(func(t *testing.T, data []byte) {
		out, err := Sanitize(data)
		if err != nil {
			// Rejecting unrecognised or malformed input is fine.
			return
		}

		// The output must be a recognisable container, and the same format as the input.
		switch {
		case isJPEG(data):
			if !isJPEG(out) {
				t.Fatalf("JPEG output is not a JPEG")
			}
		case isPNG(data):
			if !isPNG(out) {
				t.Fatalf("PNG output is not a PNG")
			}
		case isWebP(data):
			if !isWebP(out) {
				t.Fatalf("WebP output is not a WebP")
			}
		default:
			t.Fatalf("Sanitize accepted an unrecognised format")
		}

		// Sanitizing the output again must change nothing.
		second, err := Sanitize(out)
		if err != nil {
			t.Fatalf("Sanitize failed on its own output: %v", err)
		}
		if !bytes.Equal(out, second) {
			t.Fatalf("Sanitize is not idempotent")
		}
	})
}
