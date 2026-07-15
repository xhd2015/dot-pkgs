## Expected

- Root `""` has child `sub/`.
- `sub/` has child `sub/nested/`.
- `sub/nested/` has no children.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	rootChildren := resp.ScopeTree.Children[tagscope.TagScopeKey("")]
	if len(rootChildren) != 1 || rootChildren[0] != tagscope.TagScopeKey("sub/") {
		t.Fatalf("root children = %v, want [sub/]", rootChildren)
	}
	subChildren := resp.ScopeTree.Children[tagscope.TagScopeKey("sub/")]
	if len(subChildren) != 1 || subChildren[0] != tagscope.TagScopeKey("sub/nested/") {
		t.Fatalf("sub/ children = %v, want [sub/nested/]", subChildren)
	}
	nestedChildren := resp.ScopeTree.Children[tagscope.TagScopeKey("sub/nested/")]
	if nestedChildren != nil && len(nestedChildren) != 0 {
		t.Fatalf("sub/nested/ children = %v, want none", nestedChildren)
	}
}
```