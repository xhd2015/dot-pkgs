# http-proxy Test Cases

A CLI forward proxy with dynamic upstream health monitoring.

```
http-proxy --help
http-proxy --listen-port PORT --upstream-proxy URL [--fallback-direct]
```

## Test Tree

```
SETUP.md                              # Root: Request/Response, build & run binary
├── help/
│   └── show/                         # Leaf: --help prints usage
├── listen-port/
│   ├── default/                      # Leaf: default port is 7821
│   └── custom/                       # Leaf: --listen-port overrides
├── upstream/
│   ├── accessible/                   # Leaf: startup with reachable upstream
│   ├── unreachable-no-fallback/      # Leaf: startup with dead upstream, no --fallback-direct → error
│   └── unreachable-fallback/         # Leaf: startup with dead upstream, --fallback-direct → starts
└── dynamic/
    └── switch-and-back/              # Leaf: dead → live → dead upstream transitions
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | help/show | Running `--help` prints all flags and exits 0 |
| 2 | listen-port/default | Without `--listen-port`, prints "listening on :7821" |
| 3 | listen-port/custom | `--listen-port 9999` prints "listening on :9999" |
| 4 | upstream/accessible | Upstream is reachable at startup → prints "using upstream proxy" |
| 5 | upstream/unreachable-no-fallback | Upstream is dead, no `--fallback-direct` → exits non-zero with error |
| 6 | upstream/unreachable-fallback | Upstream is dead, `--fallback-direct` → prints "falling back to direct", starts listening |
| 7 | dynamic/switch-and-back | Dead upstream → starts → becomes reachable (1s) → stays reachable (5s) → becomes dead → fallback both ways |

## Running Tests

```bash
cd go-pkgs/cmd/http-proxy
doctest build tests/
doctest test tests/
```
