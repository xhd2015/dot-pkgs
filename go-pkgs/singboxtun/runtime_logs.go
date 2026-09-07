package singboxtun

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	singBoxLogFile = "sing-box.log"
)

// activeCacheDirName is the cache subdirectory used for logs and session state.
// RunTun temporarily overrides this from RunTunOptions.CacheDirName.
var activeCacheDirName = defaultCacheDirName

// packageCacheDir returns ~/Library/Caches/<CacheDirName> (or the invoking user's
// cache when started via sudo). Logs live here so a prior root run cannot leave
// cwd log files that block non-sudo use.
func packageCacheDir() (string, error) {
	base, err := processCacheBaseDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, activeCacheDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func processCacheBaseDir() (string, error) {
	if os.Geteuid() == 0 {
		if name := os.Getenv("SUDO_USER"); name != "" && name != "root" {
			u, err := user.Lookup(name)
			if err == nil {
				return homeCacheDir(u.HomeDir), nil
			}
		}
	}
	return os.UserCacheDir()
}

func homeCacheDir(home string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Caches")
	}
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir
	}
	return filepath.Join(home, ".cache")
}

func singBoxLogPath() string {
	dir, err := packageCacheDir()
	if err != nil {
		return singBoxLogFile
	}
	return filepath.Join(dir, singBoxLogFile)
}
