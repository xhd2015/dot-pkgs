# cmd-exec — running external commands

Use `github.com/xhd2015/xgo/support/cmd` for executing external
commands. It provides a fluent builder pattern that prints the command
being run and inherits stdin/stdout/stderr by default.

Install:

```bash
go get github.com/xhd2015/xgo/support/cmd@latest
```

## Quick start

```go
package main

import (
    "fmt"
    "os"

    "github.com/xhd2015/xgo/support/cmd"
)

func main() {
    // Debug mode: prints "[cmd] go version" before executing,
    // inherits stdin/stdout/stderr.
    if err := cmd.Debug().Run("go", "version"); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Running a command in a directory

```go
err := cmd.Debug().Dir("/path/to/project").Run("go", "build", "./...")
```

Chain `.Dir(...)` before `.Run(...)`. The working directory applies
only to that single invocation.

## Setting environment variables

```go
err := cmd.Debug().Env([]string{
    "GOOS=linux",
    "GOARCH=amd64",
}).Run("go", "build", "-o", "bin/app", ".")
```

## Capturing output

Use `cmd.Output` when you need stdout as a byte slice:

```go
out, err := cmd.Output("git", "rev-parse", "--show-toplevel")
// out is []byte, empty on error
```

## Suppressing output

Redirect stderr or stdout to `io.Discard`:

```go
import "io"

err := cmd.Debug().Stderr(io.Discard).Run("git", "status")
```

## Passing stdin

```go
err := cmd.Debug().Stdin(os.Stdin).Run("bash", "-c", "read -p 'Name: ' name")
```

## Non-debug mode (silent)

Use `cmd.Run` (without `.Debug()`) for quiet execution — no
"[cmd] ..." prefix is printed:

```go
err := cmd.Run("go", "build", "./...")
```

## Chaining order

Methods can be chained in any order before `.Run(...)`:

```go
cmd.Debug().Dir(wd).Env(env).Stdin(os.Stdin).Stderr(io.Discard).Run("make", "test")
```
