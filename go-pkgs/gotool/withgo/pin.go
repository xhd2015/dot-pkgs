package withgo

import "github.com/xhd2015/xgo/support/downloadgo"

// pinTable is the kool ResolveGoroot major.minor → patch map (go1.14…go1.25).
var pinTable = map[string]string{
	"go1.25": "go1.25.0",
	"go1.24": "go1.24.1",
	"go1.23": "go1.23.6",
	"go1.22": "go1.22.12",
	"go1.21": "go1.21.13",
	"go1.20": "go1.20.14",
	"go1.19": "go1.19.13",
	"go1.18": "go1.18.10",
	"go1.17": "go1.17.13",
	"go1.16": "go1.16.15",
	"go1.15": "go1.15.15",
	"go1.14": "go1.14.15",
}

// PinPatch maps a go version spelling to a pinned SDK directory name.
// Naked "1.19" matches "go1.19". Already-full patches keep a "go" prefix.
// Unknown major.minor is left unchanged but still go-prefixed.
func PinPatch(goVersion string) string {
	name := downloadgo.DirName(goVersion)
	if pin, ok := pinTable[name]; ok {
		return pin
	}
	return name
}
