# http-proxy Test Cases

Run the tests:
```sh
doctest test -v ./
```

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
│   ├── accessible/                   # Leaf: startup with reachable upstream (no --fallback-direct)
│   ├── accessible-fallback/          # Leaf: startup with reachable upstream + --fallback-direct
│   ├── unreachable-no-fallback/      # Leaf: startup with dead upstream, no --fallback-direct → warns, starts
│   └── unreachable-fallback/         # Leaf: startup with dead upstream, --fallback-direct → starts
├── dynamic/
│   ├── switch-and-back/              # Leaf: dead → live → dead upstream transitions
│   └── start-live-down-up/           # Leaf: live → dead → live upstream transitions
└── request-log/
    ├── direct/                       # Leaf: request logged as "via direct"
    └── upstream/                     # Leaf: request logged as "via upstream proxy"
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | help/show | Running `--help` prints all flags and exits 0 |
| 2 | listen-port/default | Without `--listen-port`, prints "listening on :7821" |
| 3 | listen-port/custom | `--listen-port 9999` prints "listening on :9999" |
| 4 | upstream/accessible | Upstream is reachable at startup → prints "using upstream proxy" |
| 5 | upstream/accessible-fallback | Upstream reachable + `--fallback-direct` → health check loop starts on success |
| 6 | upstream/unreachable-no-fallback | Upstream is dead, no `--fallback-direct` → prints warning, starts listening |
| 7 | upstream/unreachable-fallback | Upstream is dead, `--fallback-direct` → prints "falling back to direct", starts listening |
| 8 | dynamic/switch-and-back | Dead upstream → becomes reachable → becomes dead → fallback transitions |
| 9 | dynamic/start-live-down-up | Live upstream → goes dead → comes back live → health check detects both |
| 10 | request-log/direct | HTTP request in fallback mode logs "via direct" |
| 11 | request-log/upstream | HTTP request through upstream proxy logs "via upstream proxy" |

## Running Tests

```bash
cd go-pkgs/cmd/http-proxy
doctest build tests/
doctest test tests/
```
