# flags — sub-command dispatcher pattern

Use `StopOnFirstArg()` so top-level flags stop parsing at the first
positional argument. The remainder can then be dispatched to a
sub-command handler, which runs its own `flags.Parse` over its own
option set.

```go
package main

import (
    "fmt"
    "os"

    "github.com/xhd2015/less-gen/flags"
)

const topHelp = `
Usage: mytool [global-options] <command> [ARGS]

Commands:
  install [<dir>]    install something to <dir>
  run <script>       run a script

Global options:
  --debug        enable debug output
  -h, --help     show this help
`

func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

func run(args []string) error {
    var debug bool
    args, err := flags.Bool("--debug", &debug).
        Help("-h,--help", topHelp).
        StopOnFirstArg().
        Parse(args)
    if err != nil {
        return err
    }
    if len(args) == 0 {
        fmt.Print(topHelp)
        return nil
    }

    switch args[0] {
    case "install":
        return handleInstall(args[1:])
    case "run":
        return handleRun(args[1:])
    default:
        return fmt.Errorf("unknown command: %s", args[0])
    }
}

func handleInstall(args []string) error {
    var force bool
    args, err := flags.Bool("--force", &force).
        Help("-h,--help", `
Usage: install [--force] <dir>
`).Parse(args)
    if err != nil {
        return err
    }
    if len(args) == 0 {
        return fmt.Errorf("install requires <dir>")
    }
    fmt.Printf("installing to %s (force=%v)\n", args[0], force)
    return nil
}

func handleRun(args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("run requires a script")
    }
    fmt.Printf("running %s\n", args[0])
    return nil
}
```

## Notes

- Without `StopOnFirstArg()`, `flags` would try to interpret
  sub-command flags (e.g. `--force` after `install`) against the
  top-level spec and fail with `unrecognized flag`.
- Each handler can reuse `flags.Help(...)` for its own `--help`.
- See the `flags-parsing/types` sub-topic for the full list of
  supported target types.
