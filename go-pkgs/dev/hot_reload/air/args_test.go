package air

import (
	"reflect"
	"testing"
)

func TestBuildArgs_Defaults(t *testing.T) {
	args, err := BuildArgs(Options{
		BuildCmd:   "go build -o ./tmp/x ./cmd/x",
		Entrypoint: "./tmp/x",
		ArgsBin:    []string{"--dev", "--port", "8080"},
		ExcludeDir: []string{"tmp", "node_modules"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"--build.cmd", "go build -o ./tmp/x ./cmd/x",
		"--build.entrypoint", "./tmp/x",
		"--build.delay", "500",
		"--build.include_ext", "go",
		"--build.exclude_regex", `_test\.go`,
		"--build.exclude_unchanged", "true",
		"--build.exclude_dir", "tmp,node_modules",
		"--build.args_bin", "--dev",
		"--build.args_bin", "--port",
		"--build.args_bin", "8080",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("got %#v\nwant %#v", args, want)
	}
}

func TestBuildArgs_IncludeDir(t *testing.T) {
	falseVal := false
	args, err := BuildArgs(Options{
		BuildCmd:         "go build -o ./tmp/aiw-dev ./script/ai-workshop/dev",
		Entrypoint:       "./tmp/aiw-dev",
		IncludeDir:       []string{"ai-workshop", "script/ai-workshop"},
		ExcludeDir:       []string{"vendor", "tmp"},
		ExcludeUnchanged: &falseVal,
		DelayMs:          250,
		IncludeExt:       []string{"go", "mod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := []string{
		"--build.cmd", "go build -o ./tmp/aiw-dev ./script/ai-workshop/dev",
		"--build.entrypoint", "./tmp/aiw-dev",
		"--build.delay", "250",
		"--build.include_ext", "go,mod",
		"--build.exclude_regex", `_test\.go`,
		"--build.exclude_unchanged", "false",
		"--build.include_dir", "ai-workshop,script/ai-workshop",
		"--build.exclude_dir", "vendor,tmp",
	}
	if !reflect.DeepEqual(args, wantPrefix) {
		t.Fatalf("got %#v\nwant %#v", args, wantPrefix)
	}
}

func TestBuildArgs_RequiresFields(t *testing.T) {
	if _, err := BuildArgs(Options{Entrypoint: "./tmp/x"}); err == nil {
		t.Fatal("expected BuildCmd error")
	}
	if _, err := BuildArgs(Options{BuildCmd: "go build"}); err == nil {
		t.Fatal("expected Entrypoint error")
	}
}
