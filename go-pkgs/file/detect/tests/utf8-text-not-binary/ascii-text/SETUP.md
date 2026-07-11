# Scenario

**Feature**: plain ASCII text is not binary

```
"hello world\n" -> DetectFileType -> isBinary=false
```

## Steps

1. Write a temp file with ASCII content `hello world\n`.
2. Set `req.Path` to that file.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = writeTempFile(t, "hello.txt", []byte("hello world\n"))
	return nil
}
```
