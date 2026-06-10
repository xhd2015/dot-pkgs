# http-proxy-flex

A CLI forward HTTP proxy with dynamic upstream health monitoring and fallback.

## Usage

```bash
http-proxy-flex --help
http-proxy-flex --listen-port 7821 --upstream-proxy http://localhost:1087
http-proxy-flex --listen-port 7821 --upstream-proxy http://localhost:1087 --fallback-direct
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--listen-port` | int | `7821` | Port to listen on |
| `--upstream-proxy` | string | *(required)* | Upstream proxy URL (e.g. `http://localhost:1087`) |
| `--fallback-direct` | bool | `false` | Fall back to direct network access if upstream is unreachable |
| `-h, --help` | bool | | Show help |

## Behavior

- **Startup**: Probes `--upstream-proxy` via TCP dial. If reachable, forwards through it. If unreachable with `--fallback-direct`, falls back to direct. Without `--fallback-direct`, exits with error.
- **Health monitoring** (with `--fallback-direct`): Checks upstream every 1s. Switches between upstream proxy and direct on upstream availability changes.
- **Request logging**: One line per request — `CONNECT <host> via <upstream proxy|direct>` or `<METHOD> <url> via <upstream proxy|direct>`.

## Example

```bash
# Start proxy on port 7821, using localhost:1087 as upstream with fallback
http-proxy-flex --listen-port 7821 --upstream-proxy http://localhost:1087 --fallback-direct
```

Then configure applications:

```bash
export http_proxy=http://localhost:7821
export https_proxy=http://localhost:7821
```

## Run Tests

```bash
doctest test tests/
```
