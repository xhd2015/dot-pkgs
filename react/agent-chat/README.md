# @xhd2015/dot-pkgs-react-agent-chat

Source-only presentational agent chat: message log, bubbles, composer,
thinking, pending/retry. Consumers import TypeScript; Vite compiles it.
There is no prebuilt ESM and no transport.

```ts
import { AgentChat } from '@xhd2015/dot-pkgs-react-agent-chat'
```

## Install

Like `go.mod replace`, wrk worktrees point at the brought tree:

```json
"@xhd2015/dot-pkgs-react-agent-chat": "file:../../external/dot-pkgs/react/agent-chat"
```

Keep a Vite alias to `src/index.ts` when that folder exists so npm `file:` still hot-reloads.

## Tests

```sh
bun install
bun test
```
