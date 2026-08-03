# Scenario

**Feature**: BuildUserSpaceIndex keeps only type==0 and assigns dense 0-based indices

```
Spaces [id=3 type0, id=50 type4, id=132 type0, id=234 type0]
  -> BuildUserSpaceIndex
  -> {3:0, 132:1, 234:2}  # type4 omitted; dense over type0 only
```

## Steps

1. Phase `pure-build-index`.
2. Mixed-type Spaces list including one non-type-0 entry between type-0 Desktops.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Phase = "pure-build-index"
	req.Spaces = []SpaceInfoInput{
		{ID: 3, Type: 0},
		{ID: 50, Type: 4},
		{ID: 132, Type: 0},
		{ID: 234, Type: 0},
	}
	return nil
}
```
