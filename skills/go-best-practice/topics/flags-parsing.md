# flags — CLI flag parsing

A fluent flag parser for Go, **mainly used for CLI flag parsing**
(boolean/string/int/duration/string-slice options, with single or
multiple aliases per flag, and a built-in `--help` path).

Install:

```bash
go get github.com/xhd2015/less-gen/flags@latest
```

Example:

```go
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: myapp [options] [args...]

Options:
  --timeout DURATION  set timeout duration
  --file FILE         add files to process, can be repeated
  -v, --verbose       enable verbose output
  -h, --help          show help
`

func main() {
    var verbose bool
    var timeout time.Duration
    var files []string

    remainArgs, err := flags.Duration("--timeout", &timeout).
        StringSlice("--file", &files).
        Bool("-v,--verbose", &verbose).
        Help("-h,--help", help).
        Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("verbose:", verbose)
    fmt.Println("timeout:", timeout)
    fmt.Println("files:", files)
    fmt.Println("args:", remainArgs)
}
```

## Help options

- `Help("-h,--help", helpText)` — prints `helpText` and exits 0.
- `HelpFunc("-h,--help", fn)` — calls `fn` and exits 0.
- `HelpNoExit()` — do not exit on help; `Parse` returns `flags.ErrHelp`.

## Multiple flag names

Comma-separate aliases in the name string:

```go
flags.Bool("-v,--verbose", &verbose)
flags.String("-o,--output", &output)
```

## Sub-topics

- `flags-parsing/types` — the full list of supported target types
  (`*bool`, `*string`, `*int`/`*int64`, `*time.Duration`, `*[]string`,
  and their `**T` variants), plus how to detect "unset" flags.
- `flags-parsing/subcommand` — sub-command dispatcher pattern using
  `StopOnFirstArg`.

Reveal with:

```bash
go-best-practice-skill flags-parsing/types
go-best-practice-skill flags-parsing/subcommand
```
