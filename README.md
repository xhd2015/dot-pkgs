# dot-pkgs

A standalone shared repository for holding shared packages across different languages.

## Structure

- `go-pkgs/` — Shared Go packages
- `react/` — Shared React helpers (route prefix, API client)
- `react/diagrams/` — Sequence, process, and draw.io renderers (`@xhd2015/dot-pkgs-react-diagrams`)
- `react/agent-chat/` — Presentational agent chat (`@xhd2015/dot-pkgs-react-agent-chat`)
- `react/terminal/` — xterm view with fit sizing and the websocket PTY protocol (`@xhd2015/dot-pkgs-react-terminal`)

Install from GitHub (root `package.json` re-exports `react/diagrams`; bun/npm do not honor `#path:`):

```json
"@xhd2015/dot-pkgs-react-diagrams": "github:xhd2015/dot-pkgs#v0.0.12"
```
