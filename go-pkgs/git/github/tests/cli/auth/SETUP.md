# Scenario

**Feature**: auth failure from `ListRepos` surfaces on CLI stderr

```
# unauthenticated gh
RunCLI repo list -> EnsureAuthenticated fails -> stderr gh auth login hint
```

## Preconditions

- Auth leaf uses mock `gh api user` failure script.

## Steps

1. Descendant `Setup` configures auth-fail mock and `repo list` argv.

## Context

- Error message must contain `gh auth login` for user remediation.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Args == nil {
		req.Args = []string{"repo", "list"}
	}
	return nil
}
```