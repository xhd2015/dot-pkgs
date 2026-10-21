# @xhd2015/dot-pkgs-react-diagrams

Source-only React renderers for UML sequence diagrams, process flowcharts, and
draw.io mxfiles. Consumers import TypeScript; Vite compiles it. There is no
prebuilt ESM.

```ts
import { SequenceDiagram } from '@xhd2015/dot-pkgs-react-diagrams/sequence'
import { ProcessDiagram } from '@xhd2015/dot-pkgs-react-diagrams/process'
import { DrawioDiagram } from '@xhd2015/dot-pkgs-react-diagrams/drawio'
```

## Install

Like `go.mod replace`, wrk worktrees point at the brought tree:

```json
"@xhd2015/dot-pkgs-react-diagrams": "file:../../external/dot-pkgs/react/diagrams"
```

After this directory is on GitHub, the committed spec can be:

```json
"@xhd2015/dot-pkgs-react-diagrams": "github:xhd2015/dot-pkgs#path:react/diagrams"
```

Keep a Vite alias to `external/dot-pkgs/react/diagrams` when that folder exists so npm `file:` still hot-reloads.

## Tests

```sh
bun install
bun test
```
