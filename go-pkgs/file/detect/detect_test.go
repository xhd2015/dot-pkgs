package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	dir := t.TempDir()

	t.Run("empty file", func(t *testing.T) {
		writeFile(t, filepath.Join(dir, "empty.txt"), "")
		_, isBin, err := DetectFileType(filepath.Join(dir, "empty.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if isBin {
			t.Fatal("empty file should not be binary")
		}
	})

	t.Run("text file", func(t *testing.T) {
		writeFile(t, filepath.Join(dir, "hello.txt"), "hello world\n")
		_, isBin, err := DetectFileType(filepath.Join(dir, "hello.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if isBin {
			t.Fatal("text file should not be binary")
		}
	})

	t.Run("binary file", func(t *testing.T) {
		writeBinaryFile(t, filepath.Join(dir, "bad.bin"), 512)
		desc, isBin, err := DetectFileType(filepath.Join(dir, "bad.bin"))
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("binary file should be detected as binary")
		}
		if desc == "" {
			t.Fatal("description should not be empty for binary file")
		}
	})

	t.Run("png image", func(t *testing.T) {
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
			0x54, 0x08, 0xD7, 0x63, 0x60, 0x60, 0x60, 0x00,
			0x00, 0x00, 0x04, 0x00, 0x01, 0x23, 0x45, 0x67,
			0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
			0xAE, 0x42, 0x60, 0x82,
		}
		pngPath := filepath.Join(dir, "test.png")
		if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
			t.Fatal(err)
		}
		_, isBin, err := DetectFileType(pngPath)
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("PNG file should be detected as binary")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, _, err := DetectFileType(filepath.Join(dir, "nonexistent"))
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("utf8 with null bytes", func(t *testing.T) {
		data := append([]byte("hello"), 0x00, 0x01, 0x02)
		path := filepath.Join(dir, "nulls.bin")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		_, isBin, err := DetectFileType(path)
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("file with null bytes should be binary")
		}
	})

	t.Run("directory", func(t *testing.T) {
		subdir := filepath.Join(dir, "subdir")
		if err := os.Mkdir(subdir, 0755); err != nil {
			t.Fatal(err)
		}
		_, isBin, err := DetectFileType(subdir)
		if err != nil {
			t.Fatalf("directory should not error: %v", err)
		}
		if isBin {
			t.Fatal("directory should not be detected as binary")
		}
	})
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeBinaryFile(t *testing.T, path string, size int) {
	t.Helper()
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
