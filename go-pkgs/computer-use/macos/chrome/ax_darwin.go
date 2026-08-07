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

