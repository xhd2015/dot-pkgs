package file

import (
	"strings"
	"testing"
)

const sampleGoMod = `module example.com/m

go 1.19

require example.com/dep v1.0.0
`

func TestDiffDataKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		after             string
		wantKinds         []Kind
		without           []Kind
		emptyAfterWithout bool
	}{
		{name: "identical", after: sampleGoMod},
		{name: "comment only", after: sampleGoMod + "// local checkout note\n"},
		{
			name:              "added replace",
			after:             sampleGoMod + "\nreplace example.com/dep => ../dep\n",
			wantKinds:         []Kind{ReplaceAdded},
			without:           []Kind{ReplaceAdded},
			emptyAfterWithout: true,
		},
		{
			name:              "added extra replace",
			after:             sampleGoMod + "\nreplace example.com/other => ../other\n",
			wantKinds:         []Kind{ReplaceAdded},
			without:           []Kind{ReplaceAdded},
			emptyAfterWithout: true,
		},
		{name: "indirect bit ignored", after: `module example.com/m

go 1.19

require example.com/dep v1.0.0 // indirect
`},
		{
			name:      "require bump",
			after:     "module example.com/m\n\ngo 1.19\n\nrequire example.com/dep v1.0.1\n",
			wantKinds: []Kind{RequireChanged},
		},
		{
			name: "require bump and added replace",
			after: `module example.com/m

go 1.19

require example.com/dep v1.0.1

replace example.com/dep => ../dep
`,
			wantKinds:         []Kind{RequireChanged, ReplaceAdded},
			without:           []Kind{ReplaceAdded},
			emptyAfterWithout: false,
		},
		{
			name:      "go version",
			after:     "module example.com/m\n\ngo 1.21\n\nrequire example.com/dep v1.0.0\n",
			wantKinds: []Kind{Go},
		},
		{
			name:      "module path",
			after:     "module example.com/other\n\ngo 1.19\n\nrequire example.com/dep v1.0.0\n",
			wantKinds: []Kind{Module},
		},
		{
			name:      "extra require",
			after:     sampleGoMod + "require example.com/extra v0.0.1\n",
			wantKinds: []Kind{RequireAdded},
		},
		{
			name:      "toolchain",
			after:     sampleGoMod + "\ntoolchain go1.22.0\n",
			wantKinds: []Kind{Toolchain},
		},
		{
			name:      "exclude added",
			after:     sampleGoMod + "\nexclude example.com/bad v1.0.0\n",
			wantKinds: []Kind{ExcludeAdded},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, err := DiffData([]byte(sampleGoMod), []byte(tc.after))
			if err != nil {
				t.Fatalf("DiffData: %v", err)
			}
			got := kinds(d)
			if !kindSlicesEqual(got, tc.wantKinds) {
				t.Fatalf("kinds=%v want %v changes=%v", got, tc.wantKinds, d.Changes)
			}
			if tc.without != nil {
				stripped := d.Without(tc.without...)
				if stripped.Empty() != tc.emptyAfterWithout {
					t.Fatalf("Without(%v).Empty()=%v want %v remaining=%v", tc.without, stripped.Empty(), tc.emptyAfterWithout, stripped.Changes)
				}
			} else if !d.Empty() != (len(tc.wantKinds) > 0) {
				t.Fatalf("Empty()=%v wantKinds=%v", d.Empty(), tc.wantKinds)
			}
		})
	}
}

func TestDiffReplaceRemovedAndChanged(t *testing.T) {
	t.Parallel()
	before := sampleGoMod + "\nreplace example.com/dep => ../dep\n"
	d, err := DiffData([]byte(before), []byte(sampleGoMod))
	if err != nil {
		t.Fatal(err)
	}
	if !kindSlicesEqual(kinds(d), []Kind{ReplaceRemoved}) {
		t.Fatalf("removed: %v", d.Changes)
	}
	if d.Without(ReplaceAdded).Empty() {
		t.Fatal("removed replace must still be present after Without(ReplaceAdded)")
	}

	after := sampleGoMod + "\nreplace example.com/dep => ../other\n"
	d, err = DiffData([]byte(before), []byte(after))
	if err != nil {
		t.Fatal(err)
	}
	if !kindSlicesEqual(kinds(d), []Kind{ReplaceChanged}) {
		t.Fatalf("changed: %v", d.Changes)
	}
}

func TestDiffDataUnparseable(t *testing.T) {
	t.Parallel()
	a := []byte{0x00, 0x01, 'a'}
	_, err := DiffData(a, a)
	if err == nil {
		t.Fatal("want parse error for NUL go.mod")
	}
}

func TestParseLaxFallback(t *testing.T) {
	t.Parallel()
	f, err := Parse("go.mod", []byte("not a go.mod {{{\n"))
	if err != nil {
		t.Fatalf("ParseLax fallback: %v", err)
	}
	if f == nil {
		t.Fatal("want non-nil File")
	}
}

func TestDiffFilesNil(t *testing.T) {
	t.Parallel()
	d := DiffFiles(nil, nil)
	if !d.Empty() {
		t.Fatalf("nil vs nil: %v", d.Changes)
	}
	after, err := Parse("go.mod", []byte(sampleGoMod))
	if err != nil {
		t.Fatal(err)
	}
	d = DiffFiles(nil, after)
	if !kindSlicesEqual(kinds(d), []Kind{Module, Go, RequireAdded}) {
		t.Fatalf("nil vs sample: %v", d.Changes)
	}
}

func kinds(d Diff) []Kind {
	seen := map[Kind]struct{}{}
	var out []Kind
	for _, c := range d.Changes {
		if _, ok := seen[c.Kind]; ok {
			continue
		}
		seen[c.Kind] = struct{}{}
		out = append(out, c.Kind)
	}
	return out
}

func kindSlicesEqual(a, b []Kind) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestKindString(t *testing.T) {
	t.Parallel()
	if ReplaceAdded.String() != "ReplaceAdded" {
		t.Fatalf("got %q", ReplaceAdded.String())
	}
	if !strings.HasPrefix(Kind(0).String(), "Kind(") {
		t.Fatalf("zero kind %q", Kind(0).String())
	}
}
