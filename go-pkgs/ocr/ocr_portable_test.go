package ocr

import (
	"context"
	"strings"
	"testing"
)

func TestRecognizeFileEmptyPathPortable(t *testing.T) {
	_, err := RecognizeFile(context.Background(), "", Options{})
	if err == nil || !strings.Contains(err.Error(), "empty image path") {
		t.Fatalf("err = %v", err)
	}
}

func TestRecognizeBytesEmptyPortable(t *testing.T) {
	_, err := RecognizeBytes(context.Background(), nil, "png", Options{})
	if err == nil || !strings.Contains(err.Error(), "empty image data") {
		t.Fatalf("err = %v", err)
	}
}
