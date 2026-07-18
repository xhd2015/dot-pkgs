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

typedef uint32_t CGSConnectionID;
typedef CGSConnectionID (*main_conn_fn)(void);
typedef CFArrayRef (*copy_spaces_fn)(CGSConnectionID);

static void *skylight_handle;
static main_conn_fn p_main_conn;
static copy_spaces_fn p_copy_spaces;

static int space_cgs_init(char *errbuf, size_t errlen) {
	if (p_copy_spaces != NULL) {
		return 0;
	}
	skylight_handle = dlopen("/System/Library/PrivateFrameworks/SkyLight.framework/SkyLight", RTLD_LAZY);
	if (!skylight_handle) {
		snprintf(errbuf, errlen, "dlopen SkyLight: %s", dlerror());
		return -1;
	}
	p_main_conn = (main_conn_fn)dlsym(skylight_handle, "CGSMainConnectionID");
	if (!p_main_conn) {
		p_main_conn = (main_conn_fn)dlsym(skylight_handle, "SLSMainConnectionID");
	}
	p_copy_spaces = (copy_spaces_fn)dlsym(skylight_handle, "CGSCopyManagedDisplaySpaces");
	if (!p_copy_spaces) {
		p_copy_spaces = (copy_spaces_fn)dlsym(skylight_handle, "SLSCopyManagedDisplaySpaces");
	}
	if (!p_main_conn || !p_copy_spaces) {
		snprintf(errbuf, errlen, "missing CGS space symbols");
		return -1;
	}
	return 0;
}

// space_cgs_spaces_xml writes a CF XML property list of managed display spaces
// into *out (caller CFRelease). Returns 0 on success.
static int space_cgs_spaces_xml(CFDataRef *out, char *errbuf, size_t errlen) {
	if (space_cgs_init(errbuf, errlen) != 0) {
		return -1;
	}
	CGSConnectionID cid = p_main_conn();
	CFArrayRef arr = p_copy_spaces(cid);
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

// countDesktopsPrivateAPI uses SkyLight CGSCopyManagedDisplaySpaces without
// opening Mission Control. Requires a working WindowServer session and cgo.
func countDesktopsPrivateAPI() (int, error) {
	var data C.CFDataRef
	errbuf := (*C.char)(C.malloc(512))
	if errbuf == nil {
		return 0, fmt.Errorf("space: malloc")
	}
	defer C.free(unsafe.Pointer(errbuf))
	if C.space_cgs_spaces_xml(&data, errbuf, 512) != 0 {
		return 0, fmt.Errorf("space: private API: %s", C.GoString(errbuf))
	}
	defer C.CFRelease(C.CFTypeRef(data))

	n := C.CFDataGetLength(data)
	if n <= 0 {
		return 0, fmt.Errorf("space: private API: empty property list")
	}
	ptr := C.CFDataGetBytePtr(data)
	xml := C.GoBytes(unsafe.Pointer(ptr), C.int(n))
	return parseManagedDisplaySpacesXML(xml)
}

func parseManagedDisplaySpacesXML(xml []byte) (int, error) {
	// Convert CF XML → JSON with plutil (always available on macOS).
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
		return 0, fmt.Errorf("space: private spaces convert: %s", msg)
	}
	return parseCGSMonitorsJSON(stdout.Bytes())
}

// parseCGSMonitorsJSON counts type-0 Spaces on the first monitor with entries.
// Input is a JSON array of monitor dicts from CGSCopyManagedDisplaySpaces.
func parseCGSMonitorsJSON(raw []byte) (int, error) {
	var monitors []map[string]interface{}
	if err := json.Unmarshal(raw, &monitors); err != nil {
		var any interface{}
		if err2 := json.Unmarshal(raw, &any); err2 != nil {
			return 0, fmt.Errorf("space: parse CGS spaces JSON: %w", err)
		}
		arr, ok := any.([]interface{})
		if !ok {
			return 0, fmt.Errorf("space: parse CGS spaces JSON: expected array: %w", err)
		}
		for _, m := range arr {
			if mm, ok := m.(map[string]interface{}); ok {
				monitors = append(monitors, mm)
			}
		}
	}
	if len(monitors) == 0 {
		return 0, fmt.Errorf("space: CGS: no monitors")
	}
	var chosen map[string]interface{}
	for _, mm := range monitors {
		spaces, _ := mm["Spaces"].([]interface{})
		if len(spaces) > 0 {
			chosen = mm
			break
		}
		if chosen == nil {
			chosen = mm
		}
	}
	pseudo, err := json.Marshal(map[string]interface{}{
		"SpacesDisplayConfiguration": map[string]interface{}{
			"Management Data": map[string]interface{}{
				"Monitors": []interface{}{chosen},
			},
		},
	})
	if err != nil {
		return 0, err
	}
	return parseSpacesPlistJSON(pseudo)
}
