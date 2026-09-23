package open

import (
	"fmt"
	"strings"
)

// family says how a browser wants a new window requested.
type family int

const (
	familyUnknown family = iota
	familyChromium
	familyFirefox
	familySafari
)

// browserSpec is one row of the browser table: the macOS application name and
// how that application opens a new window.
//
// Aliases, app name, and family live on the same row on purpose. They used to
// be two independent lists (an alias switch and a Chromium-name predicate), so
// a browser could be missing from either one and silently fall through to the
// document-open form, `open -na <app> <url>`, which hands the URL to the
// running instance as a new tab in its existing window and drags that window's
// Space forward. Opera was in neither list.
type browserSpec struct {
	app     string
	family  family
	aliases []string
}

var browserTable = []browserSpec{
	{app: "Google Chrome", family: familyChromium, aliases: []string{"chrome", "google chrome", "google-chrome"}},
	{app: "Brave Browser", family: familyChromium, aliases: []string{"brave", "brave browser"}},
	{app: "Microsoft Edge", family: familyChromium, aliases: []string{"edge", "microsoft edge", "msedge"}},
	{app: "Opera", family: familyChromium, aliases: []string{"opera", "opera browser"}},
	{app: "Vivaldi", family: familyChromium, aliases: []string{"vivaldi"}},
	{app: "Chromium", family: familyChromium, aliases: []string{"chromium"}},
	{app: "Arc", family: familyChromium, aliases: []string{"arc"}},
	{app: "Firefox", family: familyFirefox, aliases: []string{"firefox", "mozilla firefox"}},
	{app: "Safari", family: familySafari, aliases: []string{"safari"}},
}

// familyHints classifies Chromium-family applications the table does not name,
// such as "Opera GX" or "Brave Browser Beta". Matching is case-sensitive so a
// name like "Marcus" is not read as Arc.
var familyHints = []struct {
	hint   string
	family family
}{
	{"Chrome", familyChromium},
	{"Chromium", familyChromium},
	{"Brave", familyChromium},
	{"Edge", familyChromium},
	{"Opera", familyChromium},
	{"Vivaldi", familyChromium},
	{"Firefox", familyFirefox},
	{"Safari", familySafari},
}

// lookupBrowser resolves a --browser alias, or a full application name, to its
// table row. Unknown names are not an error: they are passed to `open -a` as
// given, so a browser the table does not name still opens.
func lookupBrowser(name string) (browserSpec, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return browserSpec{}, false
	}
	key := strings.ToLower(trimmed)
	for _, spec := range browserTable {
		if strings.EqualFold(spec.app, trimmed) {
			return spec, true
		}
		for _, alias := range spec.aliases {
			if alias == key {
				return spec, true
			}
		}
	}
	return browserSpec{}, false
}

// browserFamily classifies an application name: the table is authoritative,
// then the hints catch unnamed variants, else unknown.
func browserFamily(app string) family {
	if spec, ok := lookupBrowser(app); ok {
		return spec.family
	}
	for _, h := range familyHints {
		if strings.Contains(app, h.hint) {
			return h.family
		}
	}
	return familyUnknown
}

// BrowserApps returns every known --browser alias, in table order, so tests can
// assert that no alias is half-registered.
func BrowserApps() []string {
	var out []string
	for _, spec := range browserTable {
		out = append(out, spec.aliases...)
	}
	return out
}

// BrowserNames returns the canonical --browser id per browser (the first alias
// of each row), which is what help text should list.
func BrowserNames() []string {
	out := make([]string, 0, len(browserTable))
	for _, spec := range browserTable {
		out = append(out, spec.aliases[0])
	}
	return out
}

// ResolveBrowserApp maps a --browser alias to a macOS application name.
// Unknown non-empty values are returned as-is for `open -a`.
func ResolveBrowserApp(name string) string {
	if spec, ok := lookupBrowser(name); ok {
		return spec.app
	}
	return strings.TrimSpace(name)
}

// BrowserArgs returns argv that opens url in a browser. newWindow requests a
// new window so the desktop does not switch Spaces to an existing one.
func BrowserArgs(pageURL, app string, newWindow bool, goos string) ([]string, error) {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return nil, fmt.Errorf("open: target is empty")
	}
	app = ResolveBrowserApp(app)
	if goos != "darwin" {
		return URLArgs(pageURL, goos)
	}
	if !newWindow {
		if app == "" {
			return URLArgs(pageURL, goos)
		}
		return AppArgs(app, pageURL, goos)
	}
	if app == "" {
		return []string{"open", "-n", pageURL}, nil
	}
	switch browserFamily(app) {
	case familySafari:
		// Safari has no usable argv switch; AppleScript is an in-app load.
		script := fmt.Sprintf("tell application %s\nmake new document with properties {URL:%s}\nactivate\nend tell",
			appleScriptString(app), appleScriptString(pageURL))
		return []string{"osascript", "-e", script}, nil
	case familyFirefox:
		return []string{"open", "-na", app, "--args", "-new-window", pageURL}, nil
	case familyChromium:
		// `-n` launches a new instance so `--args` actually reaches the
		// browser; without it open(1) activates the running app and drops the
		// arguments, which is how the URL ends up in an existing tab.
		return []string{"open", "-na", app, "--args", "--new-window", pageURL}, nil
	default:
		return []string{"open", "-na", app, pageURL}, nil
	}
}

func appleScriptString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// Browser opens url in a new browser window. app is a --browser alias or
// empty for the default handler.
func Browser(pageURL, app string) (*Result, error) {
	return BrowserConfig(pageURL, app, true, nil)
}

// BrowserConfig is Browser with NewWindow and an injectable runner.
func BrowserConfig(pageURL, app string, newWindow bool, cfg *Config) (*Result, error) {
	target := strings.TrimSpace(pageURL)
	return launch(cfg, func(goos string) ([]string, error) {
		return BrowserArgs(target, app, newWindow, goos)
	})
}
