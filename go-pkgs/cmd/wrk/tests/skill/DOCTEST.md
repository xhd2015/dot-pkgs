# wrk skill — list, show, install

## Version
0.0.2

Decision tree for `wrk skill`: early-dispatched subcommands that read the wrk
skill from `//go:embed SKILL.md` in the wrk binary and optionally install
`SKILL.md` to agent tool directories. No git checkout is required.

# DSN (Domain Specific Notion)

- **wrk CLI** — when `args[0] == "skill"`, intercept before git/worktree logic;
  `skill` is mutually exclusive with all other wrk modes/flags (`--done`,
  `--list`, create, etc.).
- **Embedded SKILL.md** — `go-pkgs/wrkcli/SKILL.md` is compiled into the wrk
  binary via `//go:embed`; no filesystem skills project lookup at runtime.
- **wrk skill list** — prints `wrk` (single line, trailing newline); no skill
  name argument.
- **wrk skill show [--header]** — prints embedded `SKILL.md` bytes, or YAML
  frontmatter only with `---` delimiters when `--header` is set; no skill name
  argument.
- **wrk skill install [flags]** — installs the embedded `SKILL.md` via
  `github.com/xhd2015/skills/install.HandleInstall` (`SkillDirName: "wrk"`,
  no `ExtraFiles`); no skill name argument.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; used for events only.

## Tree Overview

```
skill/
├── list/
│   └── basic/                    # wrk skill list → stdout wrk\n
├── show/
│   ├── basic/                    # embedded SKILL.md (marker + name: wrk)
│   ├── header/                   # --header → YAML block with delimiters
│   └── unknown-option/           # --nope → exit 1, unknown option
├── install/
│   └── dry-run-cursor/           # --cursor --dry-run → planned .cursor/skills/wrk
└── mutual-exclusion/
    └── done/                     # wrk skill list --done → exit 1
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | list/basic | `wrk skill list` → exit 0; stdout `wrk\n` |
| 2 | show/basic | `wrk skill show` → exit 0; stdout has embedded marker and `name: wrk` |
| 3 | show/header | `wrk skill show --header` → exit 0; stdout `---\nname: wrk\n---\n` |
| 4 | show/unknown-option | `wrk skill show --nope` → exit 1; stderr unknown option |
| 5 | install/dry-run-cursor | `wrk skill install --cursor --dry-run` → dry-run lines, no writes |
| 6 | mutual-exclusion/done | `wrk skill list --done` → exit 1; mutually exclusive |

## How to Run

```sh
doctest vet ./tests/skill
doctest test ./tests/skill
doctest test ./tests/skill/list/basic
doctest test ./tests/skill/show/header
doctest test ./tests/skill/install/dry-run-cursor
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd when running wrk
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)

	args := append([]string(nil), req.Args...)
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = skillWrkEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```