//go:build darwin

package space

/*
#cgo LDFLAGS: -ldl -framework CoreFoundation
#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <CoreFoundation/CoreFoundation.h>

// Window-space-index SkyLight bindings (unique names to avoid clashing with
// count_private_darwin.go C symbols in the same package).

typedef uint32_t wsi_CGSConnectionID;
typedef wsi_CGSConnectionID (*wsi_main_conn_fn)(void);
typedef CFArrayRef (*wsi_copy_managed_fn)(wsi_CGSConnectionID);
typedef CFArrayRef (*wsi_copy_spaces_for_windows_fn)(wsi_CGSConnectionID, int, CFArrayRef);

static void *wsi_skylight;
static wsi_main_conn_fn wsi_main_conn;
static wsi_copy_managed_fn wsi_copy_managed;
static wsi_copy_spaces_for_windows_fn wsi_copy_spaces_win;

static int wsi_cgs_init(char *errbuf, size_t errlen) {
	if (wsi_main_conn != NULL && wsi_copy_managed != NULL && wsi_copy_spaces_win != NULL) {
		return 0;
	}
	if (!wsi_skylight) {
		wsi_skylight = dlopen("/System/Library/PrivateFrameworks/SkyLight.framework/SkyLight", RTLD_LAZY);
		if (!wsi_skylight) {
			snprintf(errbuf, errlen, "dlopen SkyLight: %s", dlerror());
			return -1;
		}
	}
	if (!wsi_main_conn) {
		wsi_main_conn = (wsi_main_conn_fn)dlsym(wsi_skylight, "CGSMainConnectionID");
		if (!wsi_main_conn) {
			wsi_main_conn = (wsi_main_conn_fn)dlsym(wsi_skylight, "SLSMainConnectionID");
		}
	}
	if (!wsi_copy_managed) {
		wsi_copy_managed = (wsi_copy_managed_fn)dlsym(wsi_skylight, "CGSCopyManagedDisplaySpaces");
		if (!wsi_copy_managed) {
			wsi_copy_managed = (wsi_copy_managed_fn)dlsym(wsi_skylight, "SLSCopyManagedDisplaySpaces");
		}
	}
	if (!wsi_copy_spaces_win) {
		wsi_copy_spaces_win = (wsi_copy_spaces_for_windows_fn)dlsym(wsi_skylight, "CGSCopySpacesForWindows");
		if (!wsi_copy_spaces_win) {
			wsi_copy_spaces_win = (wsi_copy_spaces_for_windows_fn)dlsym(wsi_skylight, "SLSCopySpacesForWindows");
		}
	}
	if (!wsi_main_conn || !wsi_copy_managed || !wsi_copy_spaces_win) {
		snprintf(errbuf, errlen, "missing CGS window/space symbols (main=%p managed=%p spacesForWin=%p)",
			(void *)wsi_main_conn, (void *)wsi_copy_managed, (void *)wsi_copy_spaces_win);
		return -1;
	}
	return 0;
}

// wsi_managed_spaces_xml: CF XML of CGSCopyManagedDisplaySpaces (caller CFRelease *out).
static int wsi_managed_spaces_xml(CFDataRef *out, char *errbuf, size_t errlen) {
	if (wsi_cgs_init(errbuf, errlen) != 0) {
		return -1;
	}
	wsi_CGSConnectionID cid = wsi_main_conn();
	CFArrayRef arr = wsi_copy_managed(cid);
	if (!arr) {
		snprintf(errbuf, errlen, "CGSCopyManagedDisplaySpaces returned NULL");
		return -1;
	}
	CFErrorRef cfErr = NULL;
	CFDataRef data = CFPropertyListCreateData(
		kCFAllocatorDefault,
		arr,
		kCFPropertyListXMLFormat_v1_0,
		0,
		&cfErr
	);
	CFRelease(arr);
	if (!data) {
		if (cfErr) {
			CFStringRef desc = CFErrorCopyDescription(cfErr);
			if (desc) {
				CFStringGetCString(desc, errbuf, (CFIndex)errlen, kCFStringEncodingUTF8);
				CFRelease(desc);
			} else {
				snprintf(errbuf, errlen, "CFPropertyListCreateData failed");
			}
			CFRelease(cfErr);
		} else {
			snprintf(errbuf, errlen, "CFPropertyListCreateData failed");
		}
		return -1;
	}
	*out = data;
	return 0;
}

// wsi_copy_spaces_for_windows fills *out_ids (malloc'd, caller free) with space
// ids for the given window ids. *out_n is the count. mask is typically 7.
// Returns 0 on success (including empty result with out_n=0).
static int wsi_copy_spaces_for_windows(
	int mask,
	const uint64_t *window_ids,
	int n_windows,
	uint64_t **out_ids,
	int *out_n,
	char *errbuf,
	size_t errlen
) {
	if (wsi_cgs_init(errbuf, errlen) != 0) {
		return -1;
	}
	if (n_windows < 0) {
		snprintf(errbuf, errlen, "invalid window count");
		return -1;
	}

	CFMutableArrayRef winArr = CFArrayCreateMutable(kCFAllocatorDefault, n_windows, &kCFTypeArrayCallBacks);
	if (!winArr) {
		snprintf(errbuf, errlen, "CFArrayCreateMutable failed");
		return -1;
	}
	for (int i = 0; i < n_windows; i++) {
		// CGWindow numbers are 32-bit in practice; store as SInt64 for CFNumber.
		int64_t wid = (int64_t)window_ids[i];
		CFNumberRef num = CFNumberCreate(kCFAllocatorDefault, kCFNumberSInt64Type, &wid);
		if (!num) {
			CFRelease(winArr);
			snprintf(errbuf, errlen, "CFNumberCreate window id failed");
			return -1;
		}
		CFArrayAppendValue(winArr, num);
		CFRelease(num);
	}

	wsi_CGSConnectionID cid = wsi_main_conn();
	CFArrayRef spaces = wsi_copy_spaces_win(cid, mask, winArr);
	CFRelease(winArr);
	if (!spaces) {
		// NULL often means no spaces / unknown window — treat as empty list.
		*out_ids = NULL;
		*out_n = 0;
		return 0;
	}

	CFIndex count = CFArrayGetCount(spaces);
	if (count <= 0) {
		CFRelease(spaces);
		*out_ids = NULL;
		*out_n = 0;
		return 0;
	}

	uint64_t *ids = (uint64_t *)malloc((size_t)count * sizeof(uint64_t));
	if (!ids) {
		CFRelease(spaces);
		snprintf(errbuf, errlen, "malloc space ids");
		return -1;
	}

	int written = 0;
	for (CFIndex i = 0; i < count; i++) {
		CFTypeRef item = CFArrayGetValueAtIndex(spaces, i);
		if (!item) {
			continue;
		}
		uint64_t sid = 0;
		if (CFGetTypeID(item) == CFNumberGetTypeID()) {
			// Prefer unsigned / 64-bit reads.
			long long ll = 0;
			if (CFNumberGetValue((CFNumberRef)item, kCFNumberLongLongType, &ll)) {
				sid = (uint64_t)ll;
			} else {
				int32_t i32 = 0;
				if (CFNumberGetValue((CFNumberRef)item, kCFNumberSInt32Type, &i32)) {
					sid = (uint64_t)(uint32_t)i32;
				}
			}
		} else {
			// Unexpected type — skip.
			continue;
		}
		ids[written++] = sid;
	}
	CFRelease(spaces);

	if (written == 0) {
		free(ids);
		*out_ids = NULL;
		*out_n = 0;
		return 0;
	}
	*out_ids = ids;
	*out_n = written;
	return 0;
}
*/
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"unsafe"
)

// liveManagedDisplaySpacesImpl loads CGSCopyManagedDisplaySpaces via SkyLight
// and returns all monitors as DisplaySpaces (order preserved; first = primary).
func liveManagedDisplaySpacesImpl() ([]DisplaySpaces, error) {
	var data C.CFDataRef
	errbuf := (*C.char)(C.malloc(512))
	if errbuf == nil {
		return nil, fmt.Errorf("space: malloc")
	}
	defer C.free(unsafe.Pointer(errbuf))

	if C.wsi_managed_spaces_xml(&data, errbuf, 512) != 0 {
		return nil, fmt.Errorf("space: managed display spaces: %s", C.GoString(errbuf))
	}
	defer C.CFRelease(C.CFTypeRef(data))

	n := C.CFDataGetLength(data)
	if n <= 0 {
		return nil, fmt.Errorf("space: managed display spaces: empty property list")
	}
	ptr := C.CFDataGetBytePtr(data)
	xml := C.GoBytes(unsafe.Pointer(ptr), C.int(n))
	return parseManagedDisplaySpacesListXML(xml)
}

// liveCopySpacesForWindowsImpl calls CGSCopySpacesForWindows / SLSCopySpacesForWindows
// with the given mask (production uses 7).
func liveCopySpacesForWindowsImpl(mask int, windowIDs []uint64) ([]uint64, error) {
	errbuf := (*C.char)(C.malloc(512))
	if errbuf == nil {
		return nil, fmt.Errorf("space: malloc")
	}
	defer C.free(unsafe.Pointer(errbuf))

	var cWindows *C.uint64_t
	nWin := len(windowIDs)
	if nWin > 0 {
		cWindows = (*C.uint64_t)(C.malloc(C.size_t(nWin) * C.size_t(unsafe.Sizeof(C.uint64_t(0)))))
		if cWindows == nil {
			return nil, fmt.Errorf("space: malloc window ids")
		}
		defer C.free(unsafe.Pointer(cWindows))
		slice := unsafe.Slice(cWindows, nWin)
		for i, id := range windowIDs {
			slice[i] = C.uint64_t(id)
		}
	}

	var outIDs *C.uint64_t
	var outN C.int
	if C.wsi_copy_spaces_for_windows(
		C.int(mask),
		cWindows,
		C.int(nWin),
		&outIDs,
		&outN,
		errbuf,
		512,
	) != 0 {
		return nil, fmt.Errorf("space: CopySpacesForWindows: %s", C.GoString(errbuf))
	}
	if outIDs != nil {
		defer C.free(unsafe.Pointer(outIDs))
	}
	if outN <= 0 || outIDs == nil {
		return nil, nil
	}
	raw := unsafe.Slice(outIDs, int(outN))
	out := make([]uint64, int(outN))
	for i := 0; i < int(outN); i++ {
		out[i] = uint64(raw[i])
	}
	return out, nil
}

func parseManagedDisplaySpacesListXML(xml []byte) ([]DisplaySpaces, error) {
	cmd := exec.Command("plutil", "-convert", "json", "-o", "-", "-")
	cmd.Stdin = bytes.NewReader(xml)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("space: managed spaces convert: %s", msg)
	}
	return parseCGSManagedDisplaySpacesJSON(stdout.Bytes())
}

// parseCGSManagedDisplaySpacesJSON parses the JSON array from
// CGSCopyManagedDisplaySpaces into []DisplaySpaces (all monitors, order kept).
func parseCGSManagedDisplaySpacesJSON(raw []byte) ([]DisplaySpaces, error) {
	monitors, err := decodeCGSMonitorMaps(raw)
	if err != nil {
		return nil, err
	}
	if len(monitors) == 0 {
		return nil, fmt.Errorf("space: CGS: no monitors")
	}
	out := make([]DisplaySpaces, 0, len(monitors))
	for _, mm := range monitors {
		spacesRaw, _ := mm["Spaces"].([]interface{})
		spaces := make([]SpaceInfo, 0, len(spacesRaw))
		for _, s := range spacesRaw {
			sm, ok := s.(map[string]interface{})
			if !ok {
				continue
			}
			id, ok := spaceIDFromMap(sm)
			if !ok {
				continue
			}
			spaces = append(spaces, SpaceInfo{
				ID:   id,
				Type: spaceTypeFromMap(sm),
			})
		}
		out = append(out, DisplaySpaces{Spaces: spaces})
	}
	return out, nil
}

func decodeCGSMonitorMaps(raw []byte) ([]map[string]interface{}, error) {
	var monitors []map[string]interface{}
	if err := json.Unmarshal(raw, &monitors); err == nil {
		return monitors, nil
	}
	var any interface{}
	if err := json.Unmarshal(raw, &any); err != nil {
		return nil, fmt.Errorf("space: parse CGS managed spaces JSON: %w", err)
	}
	arr, ok := any.([]interface{})
	if !ok {
		return nil, fmt.Errorf("space: parse CGS managed spaces JSON: expected array")
	}
	for _, m := range arr {
		if mm, ok := m.(map[string]interface{}); ok {
			monitors = append(monitors, mm)
		}
	}
	return monitors, nil
}

// spaceIDFromMap prefers id64 (common in both plist and CGS dumps), then id.
func spaceIDFromMap(sm map[string]interface{}) (uint64, bool) {
	for _, key := range []string{"id64", "id"} {
		if v, ok := sm[key]; ok && v != nil {
			if n, ok := jsonNumberToUint64(v); ok {
				return n, true
			}
		}
	}
	return 0, false
}

func spaceTypeFromMap(sm map[string]interface{}) int {
	v, ok := sm["type"]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0
		}
		return int(i)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}

func jsonNumberToUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil || i < 0 {
			u, err2 := n.Float64()
			if err2 != nil || u < 0 {
				return 0, false
			}
			return uint64(u), true
		}
		return uint64(i), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case uint64:
		return n, true
	default:
		return 0, false
	}
}
