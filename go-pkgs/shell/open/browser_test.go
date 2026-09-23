package open

import (
	"errors"
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

func TestHTTPHandlerBundleID(t *testing.T) {
	// The shape plutil produces for the real preferences file, including a
	// non-http handler that must be ignored and a later http entry that wins.
	doc := []byte(`{"LSHandlers":[
		{"LSHandlerURLScheme":"mailto","LSHandlerRoleAll":"com.apple.mail"},
		{"LSHandlerURLScheme":"http","LSHandlerRoleAll":"com.google.Chrome"},
		{"LSHandlerURLScheme":"https","LSHandlerRoleAll":"com.brave.Browser"},
		{"LSHandlerURLScheme":"http","LSHandlerRoleAll":"com.operasoftware.opera"}
	]}`)
	got, err := httpHandlerBundleID(doc)
	if err != nil {
		t.Fatal(err)
	}
	if got != "com.operasoftware.opera" {
		t.Fatalf("got %q, want the last http handler", got)
	}
}

func TestHTTPHandlerBundleIDErrors(t *testing.T) {
	if _, err := httpHandlerBundleID([]byte(`not json`)); err == nil {
		t.Error("expected a parse error")
	}
	if _, err := httpHandlerBundleID([]byte(`{"LSHandlers":[]}`)); err == nil {
		t.Error("expected an error when no http handler is registered")
	}
	if _, err := httpHandlerBundleID([]byte(`{"LSHandlers":[{"LSHandlerURLScheme":"mailto","LSHandlerRoleAll":"com.apple.mail"}]}`)); err == nil {
		t.Error("expected an error when only non-http handlers exist")
	}
}

func TestLookupBundleIDIsCaseInsensitive(t *testing.T) {
	// LaunchServices records lowercase while the bundle declares mixed case.
	for _, id := range []string{"com.operasoftware.opera", "com.operasoftware.Opera", "COM.OPERASOFTWARE.OPERA"} {
		spec, ok := lookupBundleID(id)
		if !ok || spec.app != "Opera" || spec.family != familyChromium {
			t.Errorf("lookupBundleID(%q) = %+v, %v", id, spec, ok)
		}
	}
	if _, ok := lookupBundleID("com.example.Unknown"); ok {
		t.Error("unknown bundle id should not resolve")
	}
	if _, ok := lookupBundleID("  "); ok {
		t.Error("blank bundle id should not resolve")
	}
}

// TestBrowserConfigResolvesDefaultBrowser is the regression for a bare
// `eluc open`: with no --browser the default handler used to get
// `open -n <url>`, which hands the URL to the running browser as a tab.
func TestBrowserConfigResolvesDefaultBrowser(t *testing.T) {
	var ran []string
	cfg := &Config{
		GOOS:           "darwin",
		DefaultBrowser: func() (string, error) { return "Opera", nil },
		Run: func(args []string) (string, error) {
			ran = append([]string{}, args...)
			return "ok", nil
		},
	}
	if _, err := BrowserConfig("http://example/", "", true, cfg); err != nil {
		t.Fatal(err)
	}
	want := []string{"open", "-na", "Opera", "--args", "--new-window", "http://example/"}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("ran %#v, want %#v", ran, want)
	}
}

func TestBrowserConfigDefaultBrowserFallbacks(t *testing.T) {
	cases := []struct {
		name     string
		resolver func() (string, error)
		want     []string
	}{
		{
			name:     "unnameable default keeps the document form",
			resolver: func() (string, error) { return "", nil },
			want:     []string{"open", "-n", "http://example/"},
		},
		{
			name:     "resolver error keeps the document form",
			resolver: func() (string, error) { return "", errors.New("no plist") },
			want:     []string{"open", "-n", "http://example/"},
		},
		{
			name:     "unlisted default browser still resolves through the table",
			resolver: func() (string, error) { return "Firefox", nil },
			want:     []string{"open", "-na", "Firefox", "--args", "-new-window", "http://example/"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var ran []string
			cfg := &Config{
				GOOS:           "darwin",
				DefaultBrowser: c.resolver,
				Run: func(args []string) (string, error) {
					ran = append([]string{}, args...)
					return "ok", nil
				},
			}
			if _, err := BrowserConfig("http://example/", "", true, cfg); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(ran, c.want) {
				t.Fatalf("ran %#v, want %#v", ran, c.want)
			}
		})
	}
}

// TestBrowserConfigNamedAppSkipsDefaultResolution keeps the default lookup off
// the path when a browser was named explicitly.
func TestBrowserConfigNamedAppSkipsDefaultResolution(t *testing.T) {
	called := false
	var ran []string
	cfg := &Config{
		GOOS: "darwin",
		DefaultBrowser: func() (string, error) {
			called = true
			return "Opera", nil
		},
		Run: func(args []string) (string, error) {
			ran = append([]string{}, args...)
			return "ok", nil
		},
	}
	if _, err := BrowserConfig("http://example/", "brave", true, cfg); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("default browser must not be resolved when an app is given")
	}
	if want := []string{"open", "-na", "Brave Browser", "--args", "--new-window", "http://example/"}; !reflect.DeepEqual(ran, want) {
		t.Fatalf("ran %#v, want %#v", ran, want)
	}
}

// TestBrowserConfigNoNewWindowSkipsDefaultResolution: the `open -a` form and
// the default handler are unrelated.
func TestBrowserConfigNoNewWindowSkipsDefaultResolution(t *testing.T) {
	called := false
	cfg := &Config{
		GOOS: "darwin",
		DefaultBrowser: func() (string, error) {
			called = true
			return "Opera", nil
		},
		Run: func(args []string) (string, error) { return "ok", nil },
	}
	if _, err := BrowserConfig("http://example/", "", false, cfg); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("default browser must not be resolved when newWindow is off")
	}
}
