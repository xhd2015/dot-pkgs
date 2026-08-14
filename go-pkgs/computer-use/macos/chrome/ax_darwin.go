//go:build darwin && cgo

package chrome

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework Cocoa
#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <stdlib.h>
#include <string.h>

static int ax_trusted(void) {
	return AXIsProcessTrusted() ? 1 : 0;
}

static AXUIElementRef ax_app(pid_t pid) {
	return AXUIElementCreateApplication(pid);
}

static void ax_release(CFTypeRef r) {
	if (r) CFRelease(r);
}

// ax_copy_str copies kAXTitleAttribute or kAXDescriptionAttribute into buf.
// which: 0=title, 1=description, 2=role, 3=value(string)
static int ax_copy_str(AXUIElementRef el, int which, char *buf, size_t buflen) {
	if (!el || !buf || buflen == 0) return -1;
	buf[0] = 0;
	CFStringRef attr = kAXTitleAttribute;
	if (which == 1) attr = kAXDescriptionAttribute;
	else if (which == 2) attr = kAXRoleAttribute;
	else if (which == 3) attr = kAXValueAttribute;
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, attr, &val) != kAXErrorSuccess || !val) {
		return -1;
	}
	int ok = -1;
	if (CFGetTypeID(val) == CFStringGetTypeID()) {
		if (CFStringGetCString((CFStringRef)val, buf, (CFIndex)buflen, kCFStringEncodingUTF8)) {
			ok = 0;
		}
	}
	CFRelease(val);
	return ok;
}

static int ax_frame_center(AXUIElementRef el, double *cx, double *cy) {
	if (!el || !cx || !cy) return -1;
	CFTypeRef posRef = NULL, sizeRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXPositionAttribute, &posRef) != kAXErrorSuccess || !posRef) {
		return -1;
	}
	if (AXUIElementCopyAttributeValue(el, kAXSizeAttribute, &sizeRef) != kAXErrorSuccess || !sizeRef) {
		CFRelease(posRef);
		return -1;
	}
	CGPoint p;
	CGSize s;
	int ok = -1;
	if (AXValueGetValue((AXValueRef)posRef, kAXValueCGPointType, &p) &&
	    AXValueGetValue((AXValueRef)sizeRef, kAXValueCGSizeType, &s)) {
		*cx = p.x + s.width / 2.0;
		*cy = p.y + s.height / 2.0;
		ok = 0;
	}
	CFRelease(posRef);
	CFRelease(sizeRef);
	return ok;
}

static void quartz_click(double x, double y) {
	CGPoint pt = CGPointMake(x, y);
	CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, pt, kCGMouseButtonLeft);
	CGEventPost(kCGHIDEventTap, move);
	CFRelease(move);
	CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
	CGEventPost(kCGHIDEventTap, down);
	CFRelease(down);
	CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, pt, kCGMouseButtonLeft);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(up);
}

// Match title, description, or string value. Prefer interactive roles for click.
static int ax_name_matches(AXUIElementRef el, const char *want, int *is_interactive) {
	char title[256], desc[256], role[128], value[256];
	title[0] = desc[0] = role[0] = value[0] = 0;
	ax_copy_str(el, 0, title, sizeof(title));
	ax_copy_str(el, 1, desc, sizeof(desc));
	ax_copy_str(el, 2, role, sizeof(role));
	ax_copy_str(el, 3, value, sizeof(value));
	int match = 0;
	if ((title[0] && strcmp(title, want) == 0) ||
	    (desc[0] && strcmp(desc, want) == 0) ||
	    (value[0] && strcmp(value, want) == 0)) {
		match = 1;
	}
	if (is_interactive) {
		*is_interactive = 0;
		if (strstr(role, "CheckBox") || strstr(role, "Button") ||
		    strstr(role, "checkBox") || strstr(role, "button") ||
		    strcmp(role, "AXCheckBox") == 0 || strcmp(role, "AXButton") == 0) {
			*is_interactive = 1;
		}
	}
	return match;
}

// Two-phase DFS: first prefer interactive match (checkbox/button), then any match.
// Returns 1 if clicked, 0 if not found.
// want_interactive: 1 = only AXCheckBox/AXButton, 0 = any named match with valid frame.
static int ax_find_click_phase(AXUIElementRef el, const char *want, int depth, int *budget, int want_interactive) {
	if (!el || !want || depth < 0 || !budget || *budget <= 0) return 0;
	(*budget)--;
	int interactive = 0;
	if (ax_name_matches(el, want, &interactive)) {
		if (!want_interactive || interactive) {
			double cx = 0, cy = 0;
			if (ax_frame_center(el, &cx, &cy) == 0) {
				// Skip zero-size / offscreen-ish frames (menu ghosts).
				if (cx > 1 && cy > 1 && (interactive || !want_interactive)) {
					// Prefer interactive: only click non-interactive in phase 2.
					if (interactive || want_interactive == 0) {
						quartz_click(cx, cy);
						return 1;
					}
				}
			}
			if (interactive) {
				AXUIElementPerformAction(el, kAXPressAction);
				return 1;
			}
		}
	}
	CFTypeRef kidsRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &kidsRef) != kAXErrorSuccess || !kidsRef) {
		return 0;
	}
	int found = 0;
	if (CFGetTypeID(kidsRef) == CFArrayGetTypeID()) {
		CFArrayRef kids = (CFArrayRef)kidsRef;
		CFIndex n = CFArrayGetCount(kids);
		for (CFIndex i = 0; i < n && *budget > 0; i++) {
			AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(kids, i);
			if (ax_find_click_phase(child, want, depth - 1, budget, want_interactive) == 1) {
				found = 1;
				break;
			}
		}
	}
	CFRelease(kidsRef);
	return found;
}

static int ax_click_named_in_pid(pid_t pid, const char *want) {
	if (!want || !*want) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	// Large budget: extensions page has many nodes; Developer mode is deep/right-side.
	// Phase 1: interactive only (checkbox/button). Phase 2: any match.
	int budget = 4000;
	int r = ax_find_click_phase(app, want, 28, &budget, 1);
	if (r != 1) {
		budget = 4000;
		r = ax_find_click_phase(app, want, 28, &budget, 0);
	}
	CFRelease(app);
	return r;
}

// ax_find_named_in_pid returns 1 if an element with title/desc/value == want exists.
static int ax_find_named_phase(AXUIElementRef el, const char *want, int depth, int *budget, int want_interactive) {
	if (!el || !want || depth < 0 || !budget || *budget <= 0) return 0;
	(*budget)--;
	int interactive = 0;
	if (ax_name_matches(el, want, &interactive)) {
		if (!want_interactive || interactive) {
			double cx = 0, cy = 0;
			if (ax_frame_center(el, &cx, &cy) == 0 && cx > 1 && cy > 1) {
				return 1;
			}
			if (interactive) return 1;
		}
	}
	CFTypeRef kidsRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &kidsRef) != kAXErrorSuccess || !kidsRef) {
		return 0;
	}
	int found = 0;
	if (CFGetTypeID(kidsRef) == CFArrayGetTypeID()) {
		CFArrayRef kids = (CFArrayRef)kidsRef;
		CFIndex n = CFArrayGetCount(kids);
		for (CFIndex i = 0; i < n && *budget > 0; i++) {
			AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(kids, i);
			if (ax_find_named_phase(child, want, depth - 1, budget, want_interactive) == 1) {
				found = 1;
				break;
			}
		}
	}
	CFRelease(kidsRef);
	return found;
}

static int ax_exists_named_in_pid(pid_t pid, const char *want) {
	if (!want || !*want) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	int budget = 4000;
	int r = ax_find_named_phase(app, want, 28, &budget, 1);
	if (r != 1) {
		budget = 4000;
		r = ax_find_named_phase(app, want, 28, &budget, 0);
	}
	CFRelease(app);
	return r;
}

// ax_count_named counts every title/desc/value match under app (for multi-card detection).
static int ax_count_named_walk(AXUIElementRef el, const char *want, int depth, int *budget) {
	if (!el || !want || depth < 0 || !budget || *budget <= 0) return 0;
	(*budget)--;
	int n = 0;
	int interactive = 0;
	if (ax_name_matches(el, want, &interactive)) {
		n = 1;
	}
	CFTypeRef kidsRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &kidsRef) != kAXErrorSuccess || !kidsRef) {
		return n;
	}
	if (CFGetTypeID(kidsRef) == CFArrayGetTypeID()) {
		CFArrayRef kids = (CFArrayRef)kidsRef;
		CFIndex nKids = CFArrayGetCount(kids);
		for (CFIndex i = 0; i < nKids && *budget > 0; i++) {
			AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(kids, i);
			n += ax_count_named_walk(child, want, depth - 1, budget);
		}
	}
	CFRelease(kidsRef);
	return n;
}

// Returns match count (>=0), or -1 on AX app failure.
static int ax_count_named_in_pid(pid_t pid, const char *want) {
	if (!want || !*want) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	int budget = 8000;
	int n = ax_count_named_walk(app, want, 28, &budget);
	CFRelease(app);
	return n;
}

typedef struct {
	char version[64];
	double cx;
	double cy;
} ax_ext_card;

static int ax_is_version_like(const char *s) {
	if (!s || !*s) return 0;
	if (s[0] < '0' || s[0] > '9') return 0;
	for (const char *p = s; *p; p++) {
		if ((*p < '0' || *p > '9') && *p != '.') return 0;
	}
	return 1;
}

static int ax_text_is(const char *a, const char *b) {
	return a && b && a[0] && strcmp(a, b) == 0;
}

static int ax_is_remove_button(AXUIElementRef el) {
	char title[256], desc[256], role[128], value[256];
	title[0] = desc[0] = role[0] = value[0] = 0;
	ax_copy_str(el, 0, title, sizeof(title));
	ax_copy_str(el, 1, desc, sizeof(desc));
	ax_copy_str(el, 2, role, sizeof(role));
	ax_copy_str(el, 3, value, sizeof(value));
	int named = ax_text_is(title, "Remove") || ax_text_is(title, "Remove from Chrome") ||
		ax_text_is(desc, "Remove") || ax_text_is(desc, "Remove from Chrome") ||
		ax_text_is(value, "Remove") || ax_text_is(value, "Remove from Chrome");
	if (!named) return 0;
	return strstr(role, "Button") != NULL || strstr(role, "button") != NULL;
}

static void ax_walk_cards(AXUIElementRef el, const char *verify, int depth, int *budget,
	int *pendingName, char *pendingVer, size_t verCap, ax_ext_card *out, int maxn, int *n) {
	if (!el || !verify || depth < 0 || !budget || *budget <= 0 || !out || !n || *n >= maxn) return;
	(*budget)--;
	int interactive = 0;
	if (ax_name_matches(el, verify, &interactive)) {
		*pendingName = 1;
		pendingVer[0] = 0;
	} else if (*pendingName) {
		char title[256], desc[256], value[256];
		title[0] = desc[0] = value[0] = 0;
		ax_copy_str(el, 0, title, sizeof(title));
		ax_copy_str(el, 1, desc, sizeof(desc));
		ax_copy_str(el, 3, value, sizeof(value));
		if (ax_is_version_like(title)) {
			strncpy(pendingVer, title, verCap - 1);
			pendingVer[verCap - 1] = 0;
		} else if (ax_is_version_like(desc)) {
			strncpy(pendingVer, desc, verCap - 1);
			pendingVer[verCap - 1] = 0;
		} else if (ax_is_version_like(value)) {
			strncpy(pendingVer, value, verCap - 1);
			pendingVer[verCap - 1] = 0;
		}
	}
	if (*pendingName && ax_is_remove_button(el)) {
		double cx = 0, cy = 0;
		if (ax_frame_center(el, &cx, &cy) == 0 && cx > 1 && cy > 1) {
			strncpy(out[*n].version, pendingVer, sizeof(out[*n].version) - 1);
			out[*n].version[sizeof(out[*n].version) - 1] = 0;
			out[*n].cx = cx;
			out[*n].cy = cy;
			(*n)++;
		}
		*pendingName = 0;
		pendingVer[0] = 0;
		return;
	}
	CFTypeRef kidsRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &kidsRef) != kAXErrorSuccess || !kidsRef) {
		return;
	}
	if (CFGetTypeID(kidsRef) == CFArrayGetTypeID()) {
		CFArrayRef kids = (CFArrayRef)kidsRef;
		CFIndex nk = CFArrayGetCount(kids);
		for (CFIndex i = 0; i < nk && *budget > 0 && *n < maxn; i++) {
			AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(kids, i);
			ax_walk_cards(child, verify, depth - 1, budget, pendingName, pendingVer, verCap, out, maxn, n);
		}
	}
	CFRelease(kidsRef);
}

static int ax_collect_named_cards_in_pid(pid_t pid, const char *verify, ax_ext_card *out, int maxn) {
	if (!verify || !*verify || !out || maxn <= 0) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	int budget = 8000;
	int n = 0;
	int pendingName = 0;
	char pendingVer[64];
	pendingVer[0] = 0;
	ax_walk_cards(app, verify, 28, &budget, &pendingName, pendingVer, sizeof(pendingVer), out, maxn, &n);
	CFRelease(app);
	return n;
}

static void ax_walk_remove_centers(AXUIElementRef el, int depth, int *budget, double *xs, double *ys, int maxn, int *n) {
	if (!el || depth < 0 || !budget || *budget <= 0 || !xs || !ys || !n || *n >= maxn) return;
	(*budget)--;
	if (ax_is_remove_button(el)) {
		double cx = 0, cy = 0;
		if (ax_frame_center(el, &cx, &cy) == 0 && cx > 1 && cy > 1) {
			xs[*n] = cx;
			ys[*n] = cy;
			(*n)++;
		}
	}
	CFTypeRef kidsRef = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &kidsRef) != kAXErrorSuccess || !kidsRef) {
		return;
	}
	if (CFGetTypeID(kidsRef) == CFArrayGetTypeID()) {
		CFArrayRef kids = (CFArrayRef)kidsRef;
		CFIndex nk = CFArrayGetCount(kids);
		for (CFIndex i = 0; i < nk && *budget > 0 && *n < maxn; i++) {
			AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(kids, i);
			ax_walk_remove_centers(child, depth - 1, budget, xs, ys, maxn, n);
		}
	}
	CFRelease(kidsRef);
}

static int ax_collect_remove_centers_in_pid(pid_t pid, double *xs, double *ys, int maxn) {
	if (!xs || !ys || maxn <= 0) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	int budget = 8000;
	int n = 0;
	ax_walk_remove_centers(app, 28, &budget, xs, ys, maxn, &n);
	CFRelease(app);
	return n;
}

static int ax_front_window_frame_in_pid(pid_t pid, double *x, double *y, double *w, double *h) {
	if (!x || !y || !w || !h) return -1;
	AXUIElementRef app = ax_app(pid);
	if (!app) return -1;
	CFTypeRef winRef = NULL;
	if (AXUIElementCopyAttributeValue(app, kAXFocusedWindowAttribute, &winRef) != kAXErrorSuccess || !winRef) {
		if (winRef) CFRelease(winRef);
		winRef = NULL;
		if (AXUIElementCopyAttributeValue(app, kAXMainWindowAttribute, &winRef) != kAXErrorSuccess || !winRef) {
			if (winRef) CFRelease(winRef);
			CFRelease(app);
			return -1;
		}
	}
	AXUIElementRef win = (AXUIElementRef)winRef;
	CFTypeRef posRef = NULL, sizeRef = NULL;
	int ok = -1;
	if (AXUIElementCopyAttributeValue(win, kAXPositionAttribute, &posRef) == kAXErrorSuccess && posRef &&
	    AXUIElementCopyAttributeValue(win, kAXSizeAttribute, &sizeRef) == kAXErrorSuccess && sizeRef) {
		CGPoint p;
		CGSize s;
		if (AXValueGetValue((AXValueRef)posRef, kAXValueCGPointType, &p) &&
		    AXValueGetValue((AXValueRef)sizeRef, kAXValueCGSizeType, &s)) {
			*x = p.x;
			*y = p.y;
			*w = s.width;
			*h = s.height;
			ok = 0;
		}
	}
	if (posRef) CFRelease(posRef);
	if (sizeRef) CFRelease(sizeRef);
	CFRelease(winRef);
	CFRelease(app);
	return ok;
}
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

func axIsTrusted() bool {
	return C.ax_trusted() != 0
}

// chromePID returns the first Google Chrome main process pid (best-effort).
func chromePID(appName string) (int, error) {
	// pgrep -x doesn't work for "Google Chrome"; use AppleScript.
	app := escapeAS(appName)
	out, errOut, err := runOSAscript(context.Background(), fmt.Sprintf(`
tell application "System Events"
  try
    return unix id of process "%s" as text
  on error
    return ""
  end try
end tell
`, app), 10*time.Second)
	if err != nil {
		return 0, fmt.Errorf("chrome pid: %v %s", err, errOut)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return 0, fmt.Errorf("chrome pid: process %q not found", appName)
	}
	var pid int
	if _, err := fmt.Sscanf(out, "%d", &pid); err != nil || pid <= 0 {
		return 0, fmt.Errorf("chrome pid: bad value %q", out)
	}
	return pid, nil
}

// axClickNamed finds an AX element with title or description == name under
// Chrome and Quartz-clicks its center (same strategy as load_chrome_extension.py).
func axClickNamed(appName, name string) error {
	if !axIsTrusted() {
		return fmt.Errorf("chrome: process is not Accessibility-trusted (System Settings → Privacy & Security → Accessibility)")
	}
	pid, err := chromePID(appName)
	if err != nil {
		return err
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	r := C.ax_click_named_in_pid(C.pid_t(pid), cname)
	if r == 1 {
		return nil
	}
	if r < 0 {
		return fmt.Errorf("chrome: AX click %q failed", name)
	}
	return fmt.Errorf("chrome: AX element %q not found", name)
}

// axExistsNamed reports whether an AX element with the given name is present
// (System Events "exists button" often false-negatives on Chrome web UI).
func axExistsNamed(appName, name string) bool {
	if !axIsTrusted() {
		return false
	}
	pid, err := chromePID(appName)
	if err != nil {
		return false
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C.ax_exists_named_in_pid(C.pid_t(pid), cname) == 1
}

// axCountNamed returns how many AX nodes match name (title/desc/value).
// -1 means AX unavailable (not trusted, no Chrome pid, or app root failed).
// Used to skip expensive System Events "entire contents" when at most one card.
func axCountNamed(appName, name string) int {
	if !axIsTrusted() {
		return -1
	}
	pid, err := chromePID(appName)
	if err != nil {
		return -1
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int(C.ax_count_named_in_pid(C.pid_t(pid), cname))
}

type axExtCard struct {
	Version string
	CX, CY  float64
}

func axCollectNamedCards(appName, verifyName string) []axExtCard {
	if !axIsTrusted() {
		return nil
	}
	pid, err := chromePID(appName)
	if err != nil {
		return nil
	}
	cname := C.CString(verifyName)
	defer C.free(unsafe.Pointer(cname))
	var buf [16]C.ax_ext_card
	n := int(C.ax_collect_named_cards_in_pid(C.pid_t(pid), cname, &buf[0], 16))
	if n <= 0 {
		return nil
	}
	out := make([]axExtCard, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, axExtCard{
			Version: C.GoString(&buf[i].version[0]),
			CX:      float64(buf[i].cx),
			CY:      float64(buf[i].cy),
		})
	}
	return out
}

func axQuartzClick(x, y float64) {
	C.quartz_click(C.double(x), C.double(y))
}

func axClickTopRightRemove(appName string) bool {
	if !axIsTrusted() {
		return false
	}
	pid, err := chromePID(appName)
	if err != nil {
		return false
	}
	var xs, ys [32]C.double
	n := int(C.ax_collect_remove_centers_in_pid(C.pid_t(pid), &xs[0], &ys[0], 32))
	if n <= 0 {
		return false
	}
	bestI := 0
	var wx, wy, ww, wh C.double
	if C.ax_front_window_frame_in_pid(C.pid_t(pid), &wx, &wy, &ww, &wh) == 0 {
		upper := float64(wy) + float64(wh)*0.4
		midX := float64(wx) + float64(ww)/2
		found := -1
		bestX := 0.0
		for i := 0; i < n; i++ {
			x, y := float64(xs[i]), float64(ys[i])
			if y < upper && x > midX && (found < 0 || x > bestX) {
				found = i
				bestX = x
			}
		}
		if found >= 0 {
			bestI = found
		} else {
			for i := 1; i < n; i++ {
				if xs[i] > xs[bestI] {
					bestI = i
				}
			}
		}
	} else {
		for i := 1; i < n; i++ {
			if xs[i] > xs[bestI] {
				bestI = i
			}
		}
	}
	axQuartzClick(float64(xs[bestI]), float64(ys[bestI]))
	return true
}

