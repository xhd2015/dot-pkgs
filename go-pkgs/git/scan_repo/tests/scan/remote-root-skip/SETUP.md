# Scenario

**Feature**: a scan root under `Library/CloudStorage` yields no discovered repos

```
# remote-backed root
caller CloudStorage root -> Scan -> immediate skip -> empty slice
```

## Steps

1. Create a CloudStorage provider directory with a fake git repo inside it.
2. Set `req.Roots` to the provider directory.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	provider := cloudStorageProvider(t, root, "GoogleDrive-user@example.com")
	cloudRepo := filepath.Join(provider, "Projects", "cloud-app")
	mkdirAll(t, cloudRepo)
	fakeGitRepo(t, cloudRepo)
	req.Roots = []string{provider}
	req.Verbose = false
	return nil
}
```