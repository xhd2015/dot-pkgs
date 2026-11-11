package service

import (
	"testing"
	"time"
)

func TestBackoffDelay(t *testing.T) {
	cases := map[int]time.Duration{
		0:  0,
		-1: 0,
		1:  time.Second,
		2:  2 * time.Second,
		3:  4 * time.Second,
		4:  8 * time.Second,
		5:  16 * time.Second,
		6:  32 * time.Second,
		7:  64 * time.Second,
		8:  128 * time.Second,
		9:  256 * time.Second,
		10: RestartBackoffMax,
		20: RestartBackoffMax,
	}
	for failures, want := range cases {
		if got := BackoffDelay(failures); got != want {
			t.Fatalf("BackoffDelay(%d) = %v, want %v", failures, got, want)
		}
	}
}

func TestBackoffDelayNeverExceedsMax(t *testing.T) {
	for failures := 1; failures <= 64; failures++ {
		if got := BackoffDelay(failures); got > RestartBackoffMax {
			t.Fatalf("BackoffDelay(%d) = %v, exceeds max %v", failures, got, RestartBackoffMax)
		}
	}
}

func TestBackoffTime(t *testing.T) {
	now := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	got := BackoffTime(3, now)
	if want := now.Add(4 * time.Second); !got.Equal(want) {
		t.Fatalf("BackoffTime(3) = %v, want %v", got, want)
	}
	if got := BackoffTime(0, now); !got.Equal(now) {
		t.Fatalf("BackoffTime(0) = %v, want %v", got, now)
	}
}
