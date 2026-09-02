package rg

import "testing"

func TestTargetTripleSupported(t *testing.T) {
	t.Parallel()
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "aarch64-apple-darwin"},
		{"darwin", "amd64", "x86_64-apple-darwin"},
		{"linux", "amd64", "x86_64-unknown-linux-musl"},
		{"linux", "arm64", "aarch64-unknown-linux-musl"},
		{"windows", "amd64", "x86_64-pc-windows-msvc"},
		{"windows", "arm64", "aarch64-pc-windows-msvc"},
	}
	for _, tc := range cases {
		got, ok := targetTriple(tc.goos, tc.goarch)
		if !ok || got != tc.want {
			t.Fatalf("%s/%s: got (%q,%v) want (%q,true)", tc.goos, tc.goarch, got, ok, tc.want)
		}
	}
}

func TestTargetTripleUnsupported(t *testing.T) {
	t.Parallel()
	if _, ok := targetTriple("plan9", "amd64"); ok {
		t.Fatal("expected unsupported")
	}
	if _, err := HostTargetTriple(); err != nil {
		// host should be supported in CI/dev; if somehow not, still OK to skip assert
		t.Log("host unsupported:", err)
	}
}

func TestAssetNameAndURL(t *testing.T) {
	t.Parallel()
	name := AssetName("v15.2.0", "aarch64-apple-darwin", "darwin")
	if name != "ripgrep-15.2.0-aarch64-apple-darwin.tar.gz" {
		t.Fatalf("asset=%q", name)
	}
	url := DownloadURL("15.2.0", "x86_64-pc-windows-msvc", "windows")
	want := "https://github.com/BurntSushi/ripgrep/releases/download/15.2.0/ripgrep-15.2.0-x86_64-pc-windows-msvc.zip"
	if url != want {
		t.Fatalf("url=%q want %q", url, want)
	}
}

func TestFormatUsingNotice(t *testing.T) {
	t.Parallel()
	sel := CLI{Path: "/a/rg", Version: "15.2.0"}
	others := []CLI{
		sel,
		{Path: "/b/rg", Version: "14.1.1"},
		{Path: "/c/rg", Version: "13.0.0"},
	}
	got := FormatUsingNotice(sel, others)
	want := "using rg 15.2.0 (/a/rg); also found 14.1.1 (/b/rg), 13.0.0 (/c/rg)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	solo := FormatUsingNotice(sel, []CLI{sel})
	if solo != "using rg 15.2.0 (/a/rg)" {
		t.Fatalf("solo=%q", solo)
	}
}
