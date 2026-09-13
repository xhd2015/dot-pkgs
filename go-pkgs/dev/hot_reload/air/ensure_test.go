package air

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEnsure_AlreadyPresent(t *testing.T) {
	var stderr bytes.Buffer
	res, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			if file == "air" {
				return "/usr/local/bin/air", nil
			}
			return "", errors.New("not found")
		},
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if res.Action != "noop" || res.BinPath != "/usr/local/bin/air" {
		t.Fatalf("got %+v", res)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected quiet noop, stderr=%q", stderr.String())
	}
}

func TestEnsure_BrewSuccess(t *testing.T) {
	haveAir := false
	var ran []string
	var stderr bytes.Buffer
	res, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			switch file {
			case "air":
				if haveAir {
					return "/opt/homebrew/bin/air", nil
				}
				return "", errors.New("missing")
			case "brew":
				return "/opt/homebrew/bin/brew", nil
			default:
				return "", errors.New("missing")
			}
		},
		RunCmd: func(ctx context.Context, name string, args ...string) error {
			ran = append(ran, name+" "+joinArgs(args))
			if name == "brew" {
				haveAir = true
			}
			return nil
		},
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if res.Action != "brew" || res.BinPath != "/opt/homebrew/bin/air" {
		t.Fatalf("got %+v", res)
	}
	if len(ran) != 1 || ran[0] != "brew install go-air" {
		t.Fatalf("ran=%v", ran)
	}
}

func TestEnsure_BrewFailThenGoInstall(t *testing.T) {
	haveAir := false
	var ran []string
	res, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			switch file {
			case "air":
				if haveAir {
					return "/Users/me/go/bin/air", nil
				}
				return "", errors.New("missing")
			case "brew":
				return "/opt/homebrew/bin/brew", nil
			case "go":
				return "/usr/local/go/bin/go", nil
			default:
				return "", errors.New("missing")
			}
		},
		RunCmd: func(ctx context.Context, name string, args ...string) error {
			ran = append(ran, name+" "+joinArgs(args))
			if name == "brew" {
				return errors.New("brew failed")
			}
			if name == "go" {
				haveAir = true
			}
			return nil
		},
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if res.Action != "go_install" {
		t.Fatalf("got %+v", res)
	}
	if len(ran) != 2 {
		t.Fatalf("ran=%v", ran)
	}
}

func TestEnsure_BothFail(t *testing.T) {
	_, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			if file == "go" || file == "brew" {
				return "/bin/" + file, nil
			}
			return "", errors.New("missing")
		},
		RunCmd: func(ctx context.Context, name string, args ...string) error {
			return errors.New("fail")
		},
		Stderr: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnsure_NoBrewNoGo(t *testing.T) {
	_, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			return "", errors.New("missing")
		},
		Stderr: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if want := "brew install go-air"; !strings.Contains(err.Error(), want) {
		t.Fatalf("err=%v", err)
	}
}

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}
