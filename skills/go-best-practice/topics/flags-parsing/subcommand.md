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

## No toplevel flags

When the toplevel itself has no flags, skip `flags.Parse` entirely.
Just check the first arg, manually handle `-h`/`--help`, and
dispatch. Each sub-command parses its own flags.

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/xhd2015/less-gen/flags"
)

const topHelp = `
Usage: mytool <command> [OPTIONS]

Commands:
  install     install deny rules
  clean       remove installed rules

Run mytool <command> --help for command-specific options.
`

func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", "mytool", err)
        os.Exit(1)
    }
}

func run(args []string) error {
    if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
        fmt.Print(strings.TrimPrefix(topHelp, "\n"))
        return nil
    }

    cmd := args[0]
    cmdArgs := args[1:]
    switch cmd {
    case "install":
        return handleInstall(cmdArgs)
    case "clean":
        return handleClean(cmdArgs)
    default:
        return fmt.Errorf("unknown command: %s", cmd)
    }
}

const installHelp = `
Usage: mytool install [--dir DIR] [--dry-run]

Options:
  --dir <dir>   project directory (default: current directory)
  --dry-run     show what would be changed without writing
  -h,--help     show help
`

func handleInstall(args []string) error {
    var dirFlag *string
    var dryRun bool
    _, err := flags.String("--dir", &dirFlag).
        Bool("--dry-run", &dryRun).
        Help("-h,--help", installHelp).
        Parse(args)
    if err != nil {
        return err
    }

    dir := "."
    if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
        dir = strings.TrimSpace(*dirFlag)
    }

    if dryRun {
        fmt.Printf("Would install to %s\n", dir)
        return nil
    }

    fmt.Printf("Installed to %s\n", dir)
    return nil
}

func handleClean(args []string) error {
    var dirFlag *string
    _, err := flags.String("--dir", &dirFlag).
        Help("-h,--help", `
Usage: mytool clean [--dir DIR]
`).Parse(args)
    if err != nil {
        return err
    }

    dir := "."
    if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
        dir = strings.TrimSpace(*dirFlag)
    }

    fmt.Printf("Cleaned %s\n", dir)
    return nil
}
```

## Notes

- Without `StopOnFirstArg()`, `flags` would try to interpret
  sub-command flags (e.g. `--force` after `install`) against the
  top-level spec and fail with `unrecognized flag`.
- When the toplevel has **no** flags, skip `flags.Parse` entirely
  at the toplevel. Just check `args[0]` for `-h`/`--help` and
  dispatch raw args to sub-commands.
- Each handler can reuse `flags.Help(...)` for its own `--help`.
- See the `flags-parsing/types` sub-topic for the full list of
  supported target types.
