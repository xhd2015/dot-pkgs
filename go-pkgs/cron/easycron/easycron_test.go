package easycron

import (
	"testing"
	"time"
)

func TestParseAndNextSmoke(t *testing.T) {
	loc := time.FixedZone("UTC", 0)

	t.Run("every-1h includes now", func(t *testing.T) {
		e, err := Parse("every-1h")
		if err != nil {
			t.Fatal(err)
		}
		anchor := time.Date(2026, 8, 24, 10, 0, 0, 0, loc)
		got, ok := e.Next(anchor, anchor, loc)
		if !ok || !got.Equal(anchor) {
			t.Fatalf("Next: got %v ok=%v, want %v", got, ok, anchor)
		}
		got2, ok := e.Next(anchor, anchor.Add(time.Nanosecond), loc)
		want2 := anchor.Add(time.Hour)
		if !ok || !got2.Equal(want2) {
			t.Fatalf("Next2: got %v ok=%v, want %v", got2, ok, want2)
		}
	})

	t.Run("every-1h-at-4m snaps forward", func(t *testing.T) {
		e, err := Parse("every-1h-at-4m")
		if err != nil {
			t.Fatal(err)
		}
		anchor := time.Date(2026, 8, 24, 10, 7, 0, 0, loc)
		got, ok := e.Next(anchor, anchor, loc)
		want := time.Date(2026, 8, 24, 11, 4, 0, 0, loc)
		if !ok || !got.Equal(want) {
			t.Fatalf("Next: got %v ok=%v, want %v", got, ok, want)
		}
	})

	t.Run("until hard stop", func(t *testing.T) {
		e, err := Parse("every-5m-until-19h00m")
		if err != nil {
			t.Fatal(err)
		}
		anchor := time.Date(2026, 8, 24, 18, 50, 0, 0, loc)
		a, ok := e.Next(anchor, anchor, loc)
		if !ok || !a.Equal(anchor) {
			t.Fatalf("first: %v ok=%v", a, ok)
		}
		b, ok := e.Next(anchor, a.Add(time.Nanosecond), loc)
		wantB := time.Date(2026, 8, 24, 18, 55, 0, 0, loc)
		if !ok || !b.Equal(wantB) {
			t.Fatalf("second: %v ok=%v want %v", b, ok, wantB)
		}
		_, ok = e.Next(anchor, b.Add(time.Nanosecond), loc)
		if ok {
			t.Fatal("expected expired after 18:55")
		}
	})

	t.Run("not-within overnight", func(t *testing.T) {
		e, err := Parse("every-5m-not-within-19h00m-to-06h30m")
		if err != nil {
			t.Fatal(err)
		}
		anchor := time.Date(2026, 8, 24, 18, 55, 0, 0, loc)
		atQuiet := time.Date(2026, 8, 24, 19, 1, 0, 0, loc)
		if e.Active(atQuiet, loc) {
			t.Fatal("19:01 should be quiet")
		}
		got, ok := e.Next(anchor, atQuiet, loc)
		want := time.Date(2026, 8, 25, 6, 30, 0, 0, loc)
		// 06:30 may not be on 5m grid from anchor — relative grid from anchor.
		// anchor 18:55, interval 5m: ... 06:25, 06:30, 06:35 on Aug 25.
		if !ok {
			t.Fatal("expected a next fire after quiet")
		}
		if got.Before(want) {
			t.Fatalf("next %v should be >= %v", got, want)
		}
		if !e.Active(got, loc) {
			t.Fatalf("next %v not active", got)
		}
	})

	t.Run("offset ge interval", func(t *testing.T) {
		_, err := Parse("every-1h-at-60m")
		if err == nil {
			t.Fatal("want error")
		}
	})
}
