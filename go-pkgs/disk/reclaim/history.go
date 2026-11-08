package reclaim

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RunRecord is one line of {Home}/.clear-tmp/runs.jsonl (CLI and web share it).
type RunRecord struct {
	ID         string       `json:"id"`
	StartedAt  string       `json:"startedAt"`  // RFC3339, local offset
	FinishedAt string       `json:"finishedAt"` // RFC3339, local offset
	Source     string       `json:"source"`     // cli | web
	DryRun     bool         `json:"dryRun"`
	Command    string       `json:"command"` // equivalent CLI invocation
	Steps      []StepResult `json:"steps"`
	TotalFreed int64        `json:"totalFreed"`
	Error      string       `json:"error,omitempty"`
}

// HistoryPath is the runs log for a home root.
func HistoryPath(home string) string {
	return filepath.Join(home, ".clear-tmp", "runs.jsonl")
}

// recordRun appends the run when opts.HistoryFile is set. Best effort: a
// history write failure only warns.
func recordRun(opts Options, started time.Time, result *Result, runErr error) {
	if opts.HistoryFile == "" {
		return
	}
	rec := RunRecord{
		ID:         started.Format("20060102-150405.000"),
		StartedAt:  formatTime(started),
		FinishedAt: formatTime(time.Now()),
		Source:     opts.Source,
		DryRun:     opts.DryRun,
		Command:    CommandFor(selectedStepIDs(opts), opts.DryRun),
		Steps:      []StepResult{},
	}
	if result != nil {
		rec.Steps = result.Steps
		rec.TotalFreed = result.TotalFreed
	}
	if runErr != nil {
		rec.Error = runErr.Error()
	}
	if err := appendRunRecord(opts.HistoryFile, rec); err != nil {
		warnf(opts.Stderr, "record run history: %v", err)
	}
}

func appendRunRecord(path string, rec RunRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

// ReadRunRecords returns up to limit records, newest first. Missing file → empty.
func ReadRunRecords(path string, limit int) ([]RunRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []RunRecord{}, nil
		}
		return nil, err
	}
	defer f.Close()
	var all []RunRecord
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		var r RunRecord
		if json.Unmarshal(sc.Bytes(), &r) == nil {
			all = append(all, r)
		}
	}
	out := make([]RunRecord, 0, len(all))
	for i := len(all) - 1; i >= 0 && (limit <= 0 || len(out) < limit); i-- {
		out = append(out, all[i])
	}
	return out, sc.Err()
}
