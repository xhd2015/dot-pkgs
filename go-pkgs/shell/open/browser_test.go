package open

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveBrowserApp(t *testing.T) {
	cases := map[string]string{
		"brave":         "Brave Browser",
		"chrome":        "Google Chrome",
		"opera":         "Opera",
		"OPERA":         "Opera",
		" Opera ":       "Opera",
		"firefox":       "Firefox",
		"vivaldi":       "Vivaldi",
		"arc":           "Arc",
		"chromium":      "Chromium",
		"Safari":        "Safari",
		"edge":          "Microsoft Edge",
		"msedge":        "Microsoft Edge",
		"Brave Browser": "Brave Browser", // a full app name also resolves
		"My Browser":    "My Browser",    // unknown is passed through
		"":              "",
	}
	for in, want := range cases {
		if got := ResolveBrowserApp(in); got != want {
			t.Errorf("ResolveBrowserApp(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestBrowserArgsEveryAlias pins the argv for each alias. Asserting the whole
// argv, not just the app name, is what catches the failure that shipped: Opera
// resolved to the right kind of name but was missing from the family list, so it
// fell through to the document-open form and reused an existing tab.
func TestBrowserArgsEveryAlias(t *testing.T) {
	const url = "http://127.0.0.1:8422/projects/eluc"
	chromium := func(app string) []string {
		return []string{"open", "-na", app, "--args", "--new-window", url}
	}
	cases := map[string][]string{
		"chrome":        chromium("Google Chrome"),
		"google chrome": chromium("Google Chrome"),
		"brave":         chromium("Brave Browser"),
		"edge":          chromium("Microsoft Edge"),
		"opera":         chromium("Opera"),
		"opera browser": chromium("Opera"),
		"vivaldi":       chromium("Vivaldi"),
		"chromium":      chromium("Chromium"),
		"arc":           chromium("Arc"),
		"firefox":       {"open", "-na", "Firefox", "--args", "-new-window", url},
	}
	for in, want := range cases {
		got, err := BrowserArgs(url, in, true, "darwin")
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("BrowserArgs(%q) = %#v, want %#v", in, got, want)
		}
	}
}

// TestBrowserAliasesAllResolve is the structural guard that makes the original
// bug impossible to reintroduce: every advertised alias must resolve to a real
// application with a known family, so no browser can be half-registered.
func TestBrowserAliasesAllResolve(t *testing.T) {
	aliases := BrowserApps()
	if len(aliases) == 0 {
		t.Fatal("BrowserApps() is empty")
	}
	for _, alias := range aliases {
		spec, ok := lookupBrowser(alias)
		if !ok {
			t.Errorf("alias %q does not resolve", alias)
			continue
		}
		if strings.TrimSpace(spec.app) == "" {
			t.Errorf("alias %q resolves to an empty app name", alias)
		}
		if spec.family == familyUnknown {
			t.Errorf("alias %q (%s) has no family, so it would reuse a tab", alias, spec.app)
		}
		// Every alias must produce the new-window form, not the document form.
		got, err := BrowserArgs("http://x/", alias, true, "darwin")
		if err != nil {
			t.Errorf("%s: %v", alias, err)
			continue
		}
		if spec.family == familySafari {
			if got[0] != "osascript" {
				t.Errorf("%s: want osascript, got %#v", alias, got)
			}
			continue
		}
		if !reflect.DeepEqual(got[:4], []string{"open", "-na", spec.app, "--args"}) {
			t.Errorf("%s: %#v is not the new-window form", alias, got)
		}
	}
}

func TestBrowserArgsDarwinNewWindow(t *testing.T) {
	got, err := BrowserArgs("http://127.0.0.1:8422/", "", true, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"open", "-n", "http://127.0.0.1:8422/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	got, err = BrowserArgs("http://127.0.0.1:8422/", "brave", true, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"open", "-na", "Brave Browser", "--args", "--new-window", "http://127.0.0.1:8422/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	got, err = BrowserArgs("http://127.0.0.1:8422/projects/eluc", "safari", true, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != "osascript" || !strings.Contains(got[2], "Safari") || !strings.Contains(got[2], "http://127.0.0.1:8422/projects/eluc") {
		t.Fatalf("%v", got)
	}
}

// TestBrowserArgsUnknownAppKeepsDocumentForm pins the fallback for browsers the
// table does not know: no --new-window guess, because a non-Chromium app would
// treat that flag as a filename.
func TestBrowserArgsUnknownAppKeepsDocumentForm(t *testing.T) {
	got, err := BrowserArgs("http://127.0.0.1:8422/", "My Browser", true, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"open", "-na", "My Browser", "http://127.0.0.1:8422/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("%#v, want %#v", got, want)
	}
}

func TestBrowserFamilyHintsCatchVariants(t *testing.T) {
	cases := map[string]family{
		"Opera GX":                  familyChromium,
		"Brave Browser Beta":        familyChromium,
		"Google Chrome Canary":      familyChromium,
		"Safari Technology Preview": familySafari,
		"Mozilla Firefox":           familyFirefox,
		"TextEdit":                  familyUnknown,
	}
	for app, want := range cases {
		if got := browserFamily(app); got != want {
			t.Errorf("browserFamily(%q) = %v, want %v", app, got, want)
		}
	}
}

func TestBrowserArgsNoNewWindowUsesAppForm(t *testing.T) {
	got, err := BrowserArgs("http://127.0.0.1:8422/", "opera", false, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"open", "-a", "Opera", "http://127.0.0.1:8422/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("%#v, want %#v", got, want)
	}
}

func TestBrowserArgsRejectsEmptyTarget(t *testing.T) {
	if _, err := BrowserArgs("   ", "opera", true, "darwin"); err == nil {
		t.Fatal("expected an error for an empty target")
	}
}

func TestBrowserArgsLinuxIgnoresApp(t *testing.T) {
	got, err := BrowserArgs("http://127.0.0.1:8422/", "brave", true, "linux")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"xdg-open", "http://127.0.0.1:8422/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestBrowserConfigInjectedRunner(t *testing.T) {
	var ran []string
	cfg := &Config{
		GOOS: "darwin",
		Run: func(args []string) (string, error) {
			ran = append([]string{}, args...)
			return "ok", nil
		},
	}
	res, err := BrowserConfig("http://example/", "chrome", true, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if res.Out != "ok" {
		t.Fatalf("out=%q", res.Out)
	}
	want := []string{"open", "-na", "Google Chrome", "--args", "--new-window", "http://example/"}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("%v", ran)
	}
}
