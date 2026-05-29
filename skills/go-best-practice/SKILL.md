---
name: go-best-practice
description: >-
  Index of Go best-practice recipes (project scaffolding, CLI flag
  parsing, and more). Use when the user wants to bootstrap a new
  project or parse CLI flags in a Go program. Run the skill to reveal
  the detailed content for a topic.
---

# Go Best Practice Skill

This skill is an **index**. Run `go-best-practice <topic>` to
print the detailed recipe for a topic. Topics are organized as a
tree; address a sub-topic with a slash-separated path, e.g.
`flags-parsing/types`.

## Topics

- `kool-create` — scaffold new projects with `kool create` (react,
  go-react, frontend, server, electron)
- `flags-parsing` — CLI flag parsing with
  `github.com/xhd2015/less-gen/flags`
  - `types` — supported target types (`*bool`, `*string`, `*int`,
    `*time.Duration`, `*[]string`, and `**T` variants)
  - `subcommand` — sub-command dispatcher patterns (with `StopOnFirstArg` and no-toplevel-flags variants)

## Usage

```bash
# list top-level topics
go-best-practice

# reveal a top-level topic
go-best-practice kool-create
go-best-practice flags-parsing

# reveal a sub-topic (slash-separated path)
go-best-practice flags-parsing/types
go-best-practice flags-parsing/subcommand

# install this SKILL.md + topics into .cursor/skills/go-best-practice
go-best-practice install --cursor
```
