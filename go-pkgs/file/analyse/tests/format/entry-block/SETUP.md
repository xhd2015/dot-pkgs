# Scenario

**Feature**: entry block orders children before semantic before aggregates

```
FormatEntryBlock -> > children -> semantic lines -> git-dirs / node_modules aggregates
```

## Steps

1. Set `Mode = format-entry`.
2. Build `.codex`-style entry with children, semantic, and aggregates.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "format-entry"
	req.Entry = analyse.EntryResult{
		Name: ".codex",
		Kind: analyse.EntryKindDir,
		Children: []analyse.ChildLine{
			{Name: "sessions", SizeHuman: "12 KB"},
			{Name: "skills", SizeHuman: "4 KB"},
		},
		Semantic: []analyse.SemanticLine{
			{Key: "sessions", Count: "2", Unit: "rollouts", SizeHuman: "12 KB"},
			{Key: "skills", Count: "1", Unit: "skill", SizeHuman: "4 KB"},
		},
		Aggregates: analyse.Aggregates{
			GitRepos:        1,
			NodeModulesDirs: 2,
		},
	}
	return nil
}
```