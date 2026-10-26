# dot-pkgs

A standalone shared repository for holding shared packages across different languages.

## Structure

- `go-pkgs/` — Shared Go packages
- `react/` — Shared React helpers (route prefix, API client)
- `react/diagrams/` — Sequence, process, and draw.io renderers (`@xhd2015/dot-pkgs-react-diagrams`)
- `react/agent-chat/` — Presentational agent chat (`@xhd2015/dot-pkgs-react-agent-chat`)
- `react/terminal/` — xterm view with fit sizing and the websocket PTY protocol (`@xhd2015/dot-pkgs-react-terminal`)

Install the source-only package from a GitHub tag. The root `package.json`
re-exports diagrams and terminal; bun/npm do not honor `#path:` subdirectory
selectors. The historical `dot-pkgs-react-diagrams` package name is retained for
compatibility.

```json
"@xhd2015/dot-pkgs-react-diagrams": "github:xhd2015/dot-pkgs#v0.0.16"
```

```ts
import { TerminalView } from '@xhd2015/dot-pkgs-react-diagrams/terminal'
```

The terminal export includes its xterm runtime dependencies. Consumers should
share/dedupe React and React DOM rather than add a local `file:` dependency.
