package githook

import "testing"

func TestParseAuthorIdent(t *testing.T) {
	a, err := ParseAuthorIdent("Xxx User <xxx@xx.xx> 1760000000 +0800")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "Xxx User" || a.Email != "xxx@xx.xx" {
		t.Fatalf("unexpected author: %#v", a)
	}
}
