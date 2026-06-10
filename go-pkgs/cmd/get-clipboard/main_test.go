package main

import (
	"regexp"
	"testing"
)

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect string
	}{
		{
			name:   "png",
			data:   []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
			expect: "png",
		},
		{
			name:   "jpeg",
			data:   []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expect: "jpg",
		},
		{
			name:   "gif87a",
			data:   []byte{'G', 'I', 'F', '8', '7', 'a'},
			expect: "gif",
		},
		{
			name:   "gif89a",
			data:   []byte{'G', 'I', 'F', '8', '9', 'a'},
			expect: "gif",
		},
		{
			name:   "bmp",
			data:   []byte{'B', 'M', 0x00, 0x00},
			expect: "bmp",
		},
		{
			name:   "tiff_little_endian",
			data:   []byte{0x49, 0x49, 0x2A, 0x00},
			expect: "tiff",
		},
		{
			name:   "tiff_big_endian",
			data:   []byte{0x4D, 0x4D, 0x00, 0x2A},
			expect: "tiff",
		},
		{
			name:   "too_short",
			data:   []byte{0x00},
			expect: "",
		},
		{
			name:   "empty",
			data:   []byte{},
			expect: "",
		},
		{
			name:   "unknown_bytes",
			data:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expect: "",
		},
		{
			name:   "partial_gif_too_short",
			data:   []byte{'G', 'I', 'F', '8'},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectImageFormat(tt.data)
			if got != tt.expect {
				t.Errorf("detectImageFormat() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	data := []byte("test image data")
	name := generateFilename(data, "png")

	pattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-clipboard-[0-9a-f]{8}\.png$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("generateFilename() = %q, does not match pattern %s", name, pattern)
	}
}

func TestGenerateFilenameConsistent(t *testing.T) {
	data := []byte("consistent data")
	name1 := generateFilename(data, "jpg")
	name2 := generateFilename(data, "jpg")
	if name1 != name2 {
		t.Errorf("same data should produce same hash: %q vs %q", name1, name2)
	}
}

func TestGenerateFilenameDifferentData(t *testing.T) {
	name1 := generateFilename([]byte("data A"), "png")
	name2 := generateFilename([]byte("data B"), "png")
	if name1 == name2 {
		t.Errorf("different data should produce different names")
	}
}
