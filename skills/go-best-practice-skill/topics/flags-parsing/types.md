# flags — supported target types

Every flag method takes a pointer target. Both `*T` and `**T` are
supported (the `**T` form lets the caller distinguish "unset" from
"zero value").

| Method                              | `*T`              | `**T`               |
| ----------------------------------- | ----------------- | ------------------- |
| `Bool(names, target)`               | `*bool`           | `**bool`            |
| `String(names, target)`             | `*string`         | `**string`          |
| `Int(names, target)`                | `*int` / `*int64` | `**int` / `**int64` |
| `Duration(names, target)`           | `*time.Duration`  | `**time.Duration`   |
| `StringSlice(names, target)`        | `*[]string`       | `**[]string`        |

## Detecting "unset" with `**T`

```go
var timeout *time.Duration
remain, err := flags.Duration("--timeout", &timeout).Parse(os.Args[1:])
if err != nil { /* ... */ }
if timeout == nil {
    // --timeout was not provided
} else {
    fmt.Println("timeout =", *timeout)
}
_ = remain
```

## Repeatable slice flags

`StringSlice` appends each occurrence:

```bash
myapp --file a.txt --file b.txt --file c.txt
# files == []string{"a.txt", "b.txt", "c.txt"}
```

## Boolean flags

Bool flags may be passed bare (`--verbose`) or with an explicit value
(`--verbose=true`, `--verbose=false`).
