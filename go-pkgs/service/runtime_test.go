package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRuntimeStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "services-runtime.json")
	store := NewRuntimeStore(path)
	if store.Path() != path {
		t.Fatalf("Path() = %q, want %q", store.Path(), path)
	}
	if state := store.Load(); len(state.Processes) != 0 {
		t.Fatalf("Load() = %+v, want empty state", state)
	}
	rec := RuntimeRecord{PID: 4242, PGID: 4242, StartedAt: "2026-03-04T05:06:07Z", Command: "sleep 300"}
	if err := store.Record("svc", rec); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	state := store.Load()
	if state.Version != RuntimeStateVersion {
		t.Fatalf("Version = %d", state.Version)
	}
	if got := state.Processes["svc"]; got != rec {
		t.Fatalf("record = %+v, want %+v", got, rec)
	}
	if ids := state.IDs(); len(ids) != 1 || ids[0] != "svc" {
		t.Fatalf("IDs() = %v", ids)
	}
	if err := store.Forget("svc"); err != nil {
		t.Fatalf("Forget() error = %v", err)
	}
	if ids := store.Load().IDs(); len(ids) != 0 {
		t.Fatalf("IDs() after Forget = %v, want empty", ids)
	}
	if err := store.Forget("svc"); err != nil {
		t.Fatalf("Forget() on a missing id error = %v, want no-op", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file left behind: %v", err)
	}
}

func TestRuntimeStoreRejectsCorruptAndUnknownVersion(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"corrupt": "{not json",
		"version": `{"version":99,"processes":{"svc":{"pid":1}}}`,
		"empty":   "",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name+".json")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			state := NewRuntimeStore(path).Load()
			if state.Version != RuntimeStateVersion || len(state.Processes) != 0 {
				t.Fatalf("Load() = %+v, want an empty usable state", state)
			}
		})
	}
}

func TestRuntimeStoreIgnoresEmptyPath(t *testing.T) {
	store := NewRuntimeStore("")
	if err := store.Record("svc", RuntimeRecord{PID: 1}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := store.Forget("svc"); err != nil {
		t.Fatalf("Forget() error = %v", err)
	}
	if err := store.Save(RuntimeState{}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if state := store.Load(); len(state.Processes) != 0 {
		t.Fatalf("Load() = %+v, want an empty state", state)
	}
}

func TestRuntimeStateSortsIDs(t *testing.T) {
	state := RuntimeState{Processes: map[string]RuntimeRecord{"c": {}, "a": {}, "b": {}}}
	ids := state.IDs()
	if len(ids) != 3 || ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Fatalf("IDs() = %v", ids)
	}
}

func TestParseRecordTime(t *testing.T) {
	want := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	if got := ParseRecordTime("2026-03-04T05:06:07Z"); !got.Equal(want) {
		t.Fatalf("ParseRecordTime() = %v, want %v", got, want)
	}
	for _, value := range []string{"", "not-a-time"} {
		if got := ParseRecordTime(value); !got.IsZero() {
			t.Fatalf("ParseRecordTime(%q) = %v, want zero", value, got)
		}
	}
}
