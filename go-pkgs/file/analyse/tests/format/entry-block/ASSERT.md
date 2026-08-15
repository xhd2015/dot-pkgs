## Expected Output

```
> .codex
  > sessions                         12 KB
  > skills                           4 KB

  sessions           2 rollouts       12 KB
  skills             1 skill          4 KB
  git-dirs          1
  node_modules      2 dirs
```

## Expected

- Child `> sessions` appears before semantic `2 rollouts`.
- Child `> skills` appears before semantic `1 skill`.
- `git-dirs` and `node_modules` aggregates appear after semantic lines.

## Errors

- `err` is nil.
- Wrong section ordering.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	block := resp.EntryBlock
	if !strings.HasPrefix(block, "> .codex") {
		t.Fatalf("block should start with > .codex:\n%s", block)
	}

	childSessions := strings.Index(block, "> sessions")
	childSkills := strings.Index(block, "> skills")
	semanticSessions := strings.Index(block, "2 rollouts")
	semanticSkills := strings.Index(block, "1 skill")
	gitDirs := strings.Index(block, "git-dirs")
	nodeModules := strings.Index(block, "node_modules")

	for name, idx := range map[string]int{
		"> sessions": childSessions,
		"> skills":   childSkills,
		"2 rollouts": semanticSessions,
		"1 skill":    semanticSkills,
		"git-dirs":   gitDirs,
		"node_modules": nodeModules,
	} {
		if idx < 0 {
			t.Fatalf("block missing %q:\n%s", name, block)
		}
	}

	if childSessions > semanticSessions || childSkills > semanticSkills {
		t.Fatalf("children should precede semantic lines:\n%s", block)
	}
	if semanticSessions > gitDirs || semanticSkills > gitDirs {
		t.Fatalf("semantic lines should precede git-dirs:\n%s", block)
	}
	if gitDirs > nodeModules {
		t.Fatalf("git-dirs should precede node_modules aggregate:\n%s", block)
	}
}
```