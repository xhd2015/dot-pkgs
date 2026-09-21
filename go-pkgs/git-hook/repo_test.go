package githook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeOriginURL(t *testing.T) {
	tests := map[string]string{
		"https://git.xxx.com/team/repo.git":      "git.xxx.com/team/repo",
		"https://git.xxx.com:8443/team/repo.git": "git.xxx.com/team/repo",
		"ssh://git@git.xxx.com:2222/team/repo":   "git.xxx.com/team/repo",
		"git@git.xxx.com:team/repo.git":          "git.xxx.com/team/repo",
		"git.xxx.com:team/repo.git":              "git.xxx.com/team/repo",
		"https://Git.XXX.com/Team/Repo":          "git.xxx.com/Team/Repo",
		"https://git.xxx.com":                    "git.xxx.com",
		"/Users/me/src/repo":                     "",
		"file:///Users/me/src/repo":              "",
		"":                                       "",
	}
	for remote, want := range tests {
		if got := NormalizeOriginURL(remote); got != want {
			t.Errorf("NormalizeOriginURL(%q) = %q, want %q", remote, got, want)
		}
	}
}

func TestMatchRepoPatterns(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		patterns []string
		origin   string
		dir      string
		want     bool
		wantErr  bool
	}{
		{
			name:     "origin glob matches full URL",
			patterns: []string{"git.xxx.com/team/*"},
			origin:   "git.xxx.com/team/legacy",
			dir:      "/Users/u/work/legacy",
			want:     true,
		},
		{
			name:     "origin glob misses other org",
			patterns: []string{"git.xxx.com/team/*"},
			origin:   "git.xxx.com/other/legacy",
			dir:      "/Users/u/work/legacy",
			want:     false,
		},
		{
			name:     "bare pattern matches origin host part",
			patterns: []string{"git.xxx.com"},
			origin:   "git.xxx.com/team/legacy",
			dir:      "/Users/u/work/legacy",
			want:     true,
		},
		{
			name:     "bare pattern matches dir basename",
			patterns: []string{"legacy"},
			origin:   "",
			dir:      "/Users/u/work/legacy",
			want:     true,
		},
		{
			name:     "star crosses separators in dir path",
			patterns: []string{"*/team/*"},
			origin:   "",
			dir:      "/Users/u/Projects/team/app",
			want:     true,
		},
		{
			name:     "pattern with slash does not fall back to basename",
			patterns: []string{"work/my-repo"},
			origin:   "",
			dir:      "/Users/u/Projects/work/my-repo",
			want:     false,
		},
		{
			name:     "full dir path matches with leading star",
			patterns: []string{"*/Projects/team/app"},
			origin:   "",
			dir:      "/Users/u/Projects/team/app",
			want:     true,
		},
		{
			name:     "question mark matches one char",
			patterns: []string{"repo?"},
			origin:   "",
			dir:      "/x/repos",
			want:     true,
		},
		{
			name:     "question mark requires the char",
			patterns: []string{"repo?"},
			origin:   "",
			dir:      "/x/repo",
			want:     false,
		},
		{
			name:     "dir prefix blocks origin match",
			patterns: []string{"dir:git.xxx.com"},
			origin:   "git.xxx.com/team/legacy",
			dir:      "/Users/u/work/legacy",
			want:     false,
		},
		{
			name:     "origin prefix blocks dir match",
			patterns: []string{"origin:legacy"},
			origin:   "",
			dir:      "/Users/u/work/legacy",
			want:     false,
		},
		{
			name:     "tilde pattern matches home dir",
			patterns: []string{"~"},
			origin:   "",
			dir:      home,
			want:     true,
		},
		{
			name:     "tilde subdir glob matches",
			patterns: []string{"~/Projects/team/*"},
			origin:   "",
			dir:      home + "/Projects/team/app",
			want:     true,
		},
		{
			name:     "tilde subdir glob misses",
			patterns: []string{"~/Projects/team/*"},
			origin:   "",
			dir:      home + "/Projects/other",
			want:     false,
		},
		{
			name:     "trailing slash in dir pattern is trimmed",
			patterns: []string{"~/Projects/team/"},
			origin:   "",
			dir:      home + "/Projects/team",
			want:     true,
		},
		{
			name:     "origin matching ignores case",
			patterns: []string{"GIT.XXX.COM/TEAM/*"},
			origin:   "git.xxx.com/team/legacy",
			dir:      "",
			want:     true,
		},
		{
			name:     "dir matching is case-sensitive",
			patterns: []string{"/Users/U/Projects/*"},
			origin:   "",
			dir:      "/Users/u/Projects/x",
			want:     false,
		},
		{
			name:     "patterns are OR'ed",
			patterns: []string{"nope.example.com", "git.xxx.com"},
			origin:   "git.xxx.com/team/legacy",
			dir:      "/Users/u/work/legacy",
			want:     true,
		},
		{
			name:     "no match without origin and dir",
			patterns: []string{"git.xxx.com"},
			origin:   "",
			dir:      "",
			want:     false,
		},
		{
			name:     "empty pattern errors",
			patterns: []string{""},
			wantErr:  true,
		},
		{
			name:     "blank pattern errors",
			patterns: []string{"   "},
			wantErr:  true,
		},
		{
			name:     "empty origin prefix value errors",
			patterns: []string{"origin:"},
			wantErr:  true,
		},
		{
			name:     "empty dir prefix value errors",
			patterns: []string{"dir:"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchRepoPatterns(tt.patterns, tt.origin, tt.dir)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("MatchRepoPatterns(%q) error = nil, want error", tt.patterns)
				}
				return
			}
			if err != nil {
				t.Fatalf("MatchRepoPatterns(%q) unexpected error: %v", tt.patterns, err)
			}
			if got != tt.want {
				t.Fatalf("MatchRepoPatterns(%q, origin=%q, dir=%q) = %v, want %v", tt.patterns, tt.origin, tt.dir, got, tt.want)
			}
		})
	}
}

func TestMatchRepoPatternsSymlinkedDir(t *testing.T) {
	realDir := filepath.Join(t.TempDir(), "real-repo")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(t.TempDir(), "linked-repo")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	// The pattern names the resolved repo; the symlinked spelling must match
	// it too, so this fails if dir resolution is ever dropped.
	resolved, err := filepath.EvalSymlinks(linkDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(resolved, "real-repo") {
		t.Fatalf("resolved dir %q does not end in real-repo", resolved)
	}
	got, err := MatchRepoPatterns([]string{"*/real-repo"}, "", linkDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("dir %q: MatchRepoPatterns = false, want true", linkDir)
	}
}

func TestDomainFilterNormalizeRepoPatterns(t *testing.T) {
	filter := DomainFilter{ExcludeRepos: []string{"  git.xxx.com/team/*  "}}
	if err := filter.Normalize(); err != nil {
		t.Fatal(err)
	}
	if len(filter.ExcludeRepos) != 1 || filter.ExcludeRepos[0] != "git.xxx.com/team/*" {
		t.Fatalf("ExcludeRepos = %#v, want trimmed single pattern", filter.ExcludeRepos)
	}

	errFilter := DomainFilter{ExcludeRepos: []string{"", "x"}}
	if err := errFilter.Normalize(); err == nil {
		t.Fatal("Normalize error = nil, want empty-pattern error")
	}
}
