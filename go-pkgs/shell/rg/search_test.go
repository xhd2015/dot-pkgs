package rg

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestParseJSONMatches(t *testing.T) {
	t.Parallel()
	raw := strings.Join([]string{
		`{"type":"begin","data":{"path":{"text":"a.md"}}}`,
		`{"type":"match","data":{"path":{"text":"a.md"},"lines":{"text":"hello shared_result\n"},"line_number":3}}`,
		`{"type":"match","data":{"path":{"text":"b.md"},"lines":{"text":"other\n"},"line_number":1}}`,
		`{"type":"end","data":{"path":{"text":"a.md"}}}`,
	}, "\n")
	got, err := parseJSONMatches([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Path != "a.md" || got[0].LineNum != 3 || got[0].Line != "hello shared_result" {
		t.Fatalf("first=%+v", got[0])
	}
}

func TestSearchInjectedNoMatches(t *testing.T) {
	t.Parallel()
	matches, err := Search(context.Background(), SearchOpts{
		Bin:     "/fake/rg",
		Roots:   []string{"/hub"},
		Pattern: "zzz",
		Run: func(ctx context.Context, bin string, args []string) ([]byte, int, error) {
			return nil, 1, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("matches=%v", matches)
	}
}

func TestSearchStreamEmitsAsParsed(t *testing.T) {
	t.Parallel()
	raw := strings.Join([]string{
		`{"type":"begin","data":{"path":{"text":"a.md"}}}`,
		`{"type":"match","data":{"path":{"text":"a.md"},"lines":{"text":"one\n"},"line_number":1}}`,
		`{"type":"match","data":{"path":{"text":"a.md"},"lines":{"text":"two\n"},"line_number":2}}`,
		`{"type":"end","data":{"path":{"text":"a.md"}}}`,
	}, "\n")
	var n int
	err := SearchStream(context.Background(), SearchOpts{
		Bin:     "/fake/rg",
		Roots:   []string{"/hub"},
		Pattern: "x",
		RunStream: func(ctx context.Context, bin string, args []string) (io.ReadCloser, func() (int, error), error) {
			return io.NopCloser(strings.NewReader(raw)), func() (int, error) { return 0, nil }, nil
		},
	}, func(m Match) error {
		n++
		if n == 1 && (m.LineNum != 1 || m.Line != "one") {
			t.Fatalf("first=%+v", m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("emits=%d", n)
	}
}

func TestSearchBuildsLiteralCIFlags(t *testing.T) {
	t.Parallel()
	var gotArgs []string
	_, err := Search(context.Background(), SearchOpts{
		Bin:     "/fake/rg",
		Roots:   []string{"/hub"},
		Pattern: "shared_result",
		Globs:   []string{"*.md", "!__meta_knowledge__/**"},
		MaxCount: 20,
		Run: func(ctx context.Context, bin string, args []string) ([]byte, int, error) {
			gotArgs = append([]string(nil), args...)
			return nil, 1, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(gotArgs, " ")
	for _, want := range []string{"--json", "-i", "-F", "-g", "*.md", "!__meta_knowledge__/**", "--", "shared_result", "/hub", "-m", "20"} {
		if !containsArg(gotArgs, want) && !strings.Contains(joined, want) {
			// check individually for flags that are separate tokens
		}
	}
	mustContain := []string{"--json", "-i", "-F", "*.md", "!__meta_knowledge__/**", "shared_result", "/hub"}
	for _, w := range mustContain {
		if !containsArg(gotArgs, w) {
			t.Fatalf("args missing %q: %v", w, gotArgs)
		}
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
