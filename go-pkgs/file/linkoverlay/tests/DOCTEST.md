# file/linkoverlay — multi-layer merge with abs symlinks + explode

Classic TDD doctests for plan phase **P1**: package
`github.com/xhd2015/dot-pkgs/go-pkgs/file/linkoverlay`. Leaves are **RED** until
real merge logic replaces the scaffold stub.

`Apply` / `ApplyDirs` merge directory seeds and sparse file overlays into a
target tree. Dir layers project **top-level absolute symlinks**; later writes
**explode** intermediate symlinks so siblings stay reachable; later layers win;
within a layer **Dir then Files**.

**Out of scope for P1:** kool refactor, go.mod replace in kool, publish bump.

## Version

0.0.2

## DSN (Domain Specific Notion)

Library merge of sparse layers onto a session target using abs symlinks and
explode-on-demand (oracle: kool `seedHomeLinked` / `materializeFilesHomeLinked`).

### Participants

- **Caller** — owns a merge **target** directory (session root) and supplies
  ordered **layers**.
- **`Apply(target, layers...)`** — merges left→right into target; later layers
  win. Each layer may have **Dir** and/or **Files**.
- **`ApplyDirs(target, dirs...)`** — shortcut for Dir-only layers in order.
- **Dir layer** — for each top-level name under Dir (including dots; skip `.` /
  `..`), create absolute symlink `target/name → abs(Dir/name)`.
- **Files overlay** — write each leaf under target with mode (0 → `0o644`);
  remove existing first (including seed symlink — no write-through into base).
- **Explode** — when an intermediate path in target is a symlink and a later op
  needs a directory there: unlink → mkdir → re-link children from the symlink’s
  target directory (absolute targets). Missing intermediate → mkdir only.
  Wrong type mid-path → replace with empty dir (no content merge).
- **Path gate** — reject `..` segments and absolute overlay paths; must stay
  under target.

### Behaviors

- Empty layer (no Dir, no Files) and empty `ApplyDirs` → success no-op.
- Same layer: Dir fully applied, then Files (Files beat same-layer Dir).
- Cross-layer: later wins on conflicting leaf paths; earlier base files on disk
  stay unchanged.
- Explode preserves siblings under a former seed symlink (e.g. `.config/other`
  still readable after writing `.config/tool/...`).
- Replacing a seeded leaf symlink must not mutate the base file content.

## Decision Tree

```
file/linkoverlay/tests/
├── seed/                              # Dir seeding / ApplyDirs
│   ├── apply-dirs-disjoint/           # two bases, disjoint tops + dot entry
│   └── empty-apply-dirs/              # ApplyDirs with no dirs → no-op
├── merge-order/                       # later wins / same-layer Dir→Files
│   ├── later-layer-wins/              # two Dir layers, same top leaf → later
│   ├── same-layer-dir-then-files/     # one layer: Dir A then Files B → B
│   └── files-last-multi/              # Dir A + Dir B + Files → files win
├── explode/                           # intermediate symlink explode
│   └── intermediate-symlink/          # .config/other + .config/tool both ok
├── leaf-safety/                       # no write-through into base
│   └── no-write-through/              # replace seeded leaf; base unchanged
└── path-reject/                       # overlay path safety
    ├── dotdot-path/                   # Path with .. → error
    └── absolute-file-path/            # absolute File.Path → error
```

Parameter ranking (most → least significant):

1. **Contract area** — seed / merge-order / explode / leaf-safety / path-reject
2. **API shape** — ApplyDirs vs Apply layers (Dir / Files mix)
3. **Conflict / path shape** — disjoint vs same-leaf, nested intermediate, invalid path

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `seed/apply-dirs-disjoint` | ApplyDirs | Two bases; disjoint tops + `.config` as abs symlinks | RED |
| 2 | `seed/empty-apply-dirs` | ApplyDirs | No dirs → success; target empty | RED |
| 3 | `merge-order/later-layer-wins` | Apply | Two Dir layers same leaf → later content; earlier base intact | RED |
| 4 | `merge-order/same-layer-dir-then-files` | Apply | One layer Dir then Files same path → Files content | RED |
| 5 | `merge-order/files-last-multi` | Apply | Dir A + Dir B + Files pack → Files beat dirs | RED |
| 6 | `explode/intermediate-symlink` | Apply | Explode `.config`; sibling + new path both readable | RED |
| 7 | `leaf-safety/no-write-through` | Apply | Replace seeded leaf symlink; base file content unchanged | RED |
| 8 | `path-reject/dotdot-path` | Apply | `File.Path` with `..` → error (not "not implemented") | RED |
| 9 | `path-reject/absolute-file-path` | Apply | Absolute `File.Path` → error (not "not implemented") | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./file/linkoverlay/tests
doctest test -v ./file/linkoverlay/tests
doctest test -v ./file/linkoverlay/tests/seed/apply-dirs-disjoint
doctest test -v ./file/linkoverlay/tests/explode/intermediate-symlink
```

Classic TDD: expect **RED** (assert failures against stub `not implemented`, or
real bugs until implementer lands merge logic).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/linkoverlay"
)

// FileSpec is a fixture file for base dirs or sparse overlays.
type FileSpec struct {
	Path    string
	Mode    uint32
	Content string
}

// LayerSpec describes one Apply layer: optional fixture Dir + sparse Files.
// Setup materializes DirRel under WorkingDir and fills BaseFiles into it.
type LayerSpec struct {
	// DirRel is relative to WorkingDir. Non-empty → create that base dir and
	// set Layer.Dir to its absolute path after writing BaseFiles.
	DirRel    string
	BaseFiles []FileSpec
	// Files become Layer.Files (sparse overlay applied after Dir in the layer).
	Files []FileSpec
}

// Request is filled root→leaf. Either UseApplyDirs or Layers is used by Run.
type Request struct {
	WorkingDir string // abs temp workspace (root Setup)
	TargetRel  string // relative to WorkingDir; default "target"

	// UseApplyDirs selects ApplyDirs(target, dirs...) using DirsRel order.
	UseApplyDirs bool
	// DirsRel are WorkingDir-relative base dirs (materialized from matching LayerSpecs).
	// When UseApplyDirs, leaves usually set Layers for fixtures and DirsRel for order;
	// if DirsRel empty, Run uses each non-empty Layer.DirRel in order.
	DirsRel []string

	Layers []LayerSpec
}

// Response observes the merge target path. Assertions inspect the filesystem.
type Response struct {
	Target string // absolute path of merge target
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()

	if req.WorkingDir == "" {
		t.Fatal("WorkingDir not set by Setup")
	}
	targetRel := req.TargetRel
	if targetRel == "" {
		targetRel = "target"
	}
	target := filepath.Join(req.WorkingDir, targetRel)
	resp := &Response{Target: target}

	if req.UseApplyDirs {
		dirs := make([]string, 0)
		if len(req.DirsRel) > 0 {
			for _, rel := range req.DirsRel {
				dirs = append(dirs, filepath.Join(req.WorkingDir, rel))
			}
		} else {
			for _, layer := range req.Layers {
				if layer.DirRel != "" {
					dirs = append(dirs, filepath.Join(req.WorkingDir, layer.DirRel))
				}
			}
		}
		err := linkoverlay.ApplyDirs(target, dirs...)
		return resp, err
	}

	layers := make([]linkoverlay.Layer, 0, len(req.Layers))
	for _, spec := range req.Layers {
		layer := linkoverlay.Layer{}
		if spec.DirRel != "" {
			layer.Dir = filepath.Join(req.WorkingDir, spec.DirRel)
		}
		if len(spec.Files) > 0 {
			layer.Files = make([]linkoverlay.File, 0, len(spec.Files))
			for _, f := range spec.Files {
				layer.Files = append(layer.Files, linkoverlay.File{
					Path:    f.Path,
					Mode:    f.Mode,
					Content: []byte(f.Content),
				})
			}
		}
		layers = append(layers, layer)
	}
	err := linkoverlay.Apply(target, layers...)
	return resp, err
}
```
