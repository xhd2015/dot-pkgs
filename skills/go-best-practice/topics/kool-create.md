# kool create — scaffold a new project

Install the `kool` CLI:

```bash
go install github.com/xhd2015/kool@latest
```

Scaffold a new project from an embedded template:

```bash
kool create <TEMPLATE> <project-name>
```

## Available templates

- `react` — a new react project
- `go-react` — go backend + react frontend
- `frontend` — react-only frontend
- `server` — go-only backend
- `electron` — electron frontend

## Examples

```bash
# react-only frontend (runs `bun install` automatically)
kool create frontend my-project

# go-only backend (renames go.mod.template -> go.mod,
# rewrites module path from the git remote when possible)
kool create server my-project

# go backend + react frontend in one repo
kool create go-react my-project

# plain react project
kool create react my-project

# electron app (runs `npm install` automatically)
kool create electron my-project
```

## After scaffolding

- `frontend`: `cd my-project && bun watch`
- `electron`: `cd my-project && npm run dev`
- `server`: a `main.go` is created with the module path set to your git
  remote (e.g. `github.com/you/repo/my-project`), or to the default
  `github.com/xhd2015/kool/tools/create/server_template` if no remote
  is found.

Show full help:

```bash
kool create --help
```
