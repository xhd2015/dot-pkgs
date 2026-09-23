package open

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	app      string
	bundleID string
	family   family
	aliases  []string
}

var browserTable = []browserSpec{
	{app: "Google Chrome", bundleID: "com.google.Chrome", family: familyChromium, aliases: []string{"chrome", "google chrome", "google-chrome"}},
	{app: "Brave Browser", bundleID: "com.brave.Browser", family: familyChromium, aliases: []string{"brave", "brave browser"}},
	{app: "Microsoft Edge", bundleID: "com.microsoft.edgemac", family: familyChromium, aliases: []string{"edge", "microsoft edge", "msedge"}},
	{app: "Opera", bundleID: "com.operasoftware.Opera", family: familyChromium, aliases: []string{"opera", "opera browser"}},
	{app: "Vivaldi", bundleID: "com.vivaldi.Vivaldi", family: familyChromium, aliases: []string{"vivaldi"}},
	{app: "Chromium", bundleID: "org.chromium.Chromium", family: familyChromium, aliases: []string{"chromium"}},
	{app: "Arc", bundleID: "company.thebrowser.Browser", family: familyChromium, aliases: []string{"arc"}},
	{app: "Firefox", bundleID: "org.mozilla.firefox", family: familyFirefox, aliases: []string{"firefox", "mozilla firefox"}},
	{app: "Safari", bundleID: "com.apple.Safari", family: familySafari, aliases: []string{"safari"}},
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

// lookupBundleID resolves a CFBundleIdentifier, case-insensitively: the
// LaunchServices preference records "com.operasoftware.opera" while the bundle
// itself declares "com.operasoftware.Opera".
func lookupBundleID(bundleID string) (browserSpec, bool) {
	key := strings.ToLower(strings.TrimSpace(bundleID))
	if key == "" {
		return browserSpec{}, false
	}
	for _, spec := range browserTable {
		if spec.bundleID != "" && strings.ToLower(spec.bundleID) == key {
			return spec, true
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

// launchServicesPlist is the per-user preference that records the system's URL
// scheme handlers.
const launchServicesPlist = "Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist"

// httpHandlerBundleID pulls the bundle identifier registered for the http
// scheme out of a LaunchServices preferences document converted to JSON.
// Later entries override earlier ones, so the last match wins.
func httpHandlerBundleID(plistJSON []byte) (string, error) {
	var doc struct {
		LSHandlers []struct {
			LSHandlerURLScheme string `json:"LSHandlerURLScheme"`
			LSHandlerRoleAll   string `json:"LSHandlerRoleAll"`
		} `json:"LSHandlers"`
	}
	if err := json.Unmarshal(plistJSON, &doc); err != nil {
		return "", fmt.Errorf("open: parse LaunchServices preferences: %w", err)
	}
	found := ""
	for _, h := range doc.LSHandlers {
		if h.LSHandlerURLScheme == "http" && h.LSHandlerRoleAll != "" {
			found = h.LSHandlerRoleAll
		}
	}
	if found == "" {
		return "", fmt.Errorf("open: no http handler in LaunchServices preferences")
	}
	return found, nil
}

// DefaultBrowserApp returns the application name of the system default http
// handler, or "" when it cannot be named.
//
// It is impure: it reads the user's LaunchServices preferences (through plutil)
// and maps the bundle identifier through the browser table. A default browser
// the table does not list yields "", which leaves the caller on the plain
// document-open form. Only darwin has this notion.
func DefaultBrowserApp() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, filepath.FromSlash(launchServicesPlist))
	out, err := exec.Command("plutil", "-convert", "json", "-o", "-", path).Output()
	if err != nil {
		return "", fmt.Errorf("open: read LaunchServices preferences: %w", err)
	}
	bundleID, err := httpHandlerBundleID(out)
	if err != nil {
		return "", err
	}
	spec, ok := lookupBundleID(bundleID)
	if !ok {
		return "", nil
	}
	return spec.app, nil
}

func appleScriptString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// Browser opens url in a new browser window. app is a --browser alias or
// empty for the default handler, which is resolved so that it too can be asked
// for a new window instead of a reused tab.
func Browser(pageURL, app string) (*Result, error) {
	return BrowserConfig(pageURL, app, true, nil)
}

// BrowserConfig is Browser with NewWindow and an injectable runner. With an
// empty app and newWindow on darwin it resolves the system default browser
// first, so the default handler gets the same new-window argv as a named one;
// when the default cannot be named the plain `open -n <url>` form is kept.
func BrowserConfig(pageURL, app string, newWindow bool, cfg *Config) (*Result, error) {
	target := strings.TrimSpace(pageURL)
	if app == "" && newWindow && target != "" && cfg.goos() == "darwin" {
		if name, err := cfg.defaultBrowser()(); err == nil && name != "" {
			app = name
		}
	}
	return launch(cfg, func(goos string) ([]string, error) {
		return BrowserArgs(target, app, newWindow, goos)
	})
}
