## Expected

- `.codex` entry exists with semantic lines.
- `sessions` semantic: `Count == "2"`, `Unit == "rollouts"`.
- `skills` semantic: `Count == "1"`, `Unit == "skill"`.
- `plugins` semantic: `Count == "0"`, `Unit == "plugins"`.
- Summary: `HasCodex == true`, `CodexSessions == 2`, `CodexSkills == 1`.
- Children include `sessions` and `skills` before semantic counts are checked via fields.

## Errors

- `err` is nil.
- Wrong rollout/skill/plugin counts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	codex := findEntry(t, resp.Entries, ".codex")

	sessions, ok := semanticByKey(codex, "sessions")
	if !ok {
		t.Fatal("missing sessions semantic line")
	}
	if sessions.Count != "2" || sessions.Unit != "rollouts" {
		t.Fatalf("sessions semantic = %+v, want 2 rollouts", sessions)
	}

	skills, ok := semanticByKey(codex, "skills")
	if !ok {
		t.Fatal("missing skills semantic line")
	}
	if skills.Count != "1" || skills.Unit != "skill" {
		t.Fatalf("skills semantic = %+v, want 1 skill", skills)
	}

	plugins, ok := semanticByKey(codex, "plugins")
	if !ok {
		t.Fatal("missing plugins semantic line")
	}
	if plugins.Count != "0" || plugins.Unit != "plugins" {
		t.Fatalf("plugins semantic = %+v, want 0 plugins", plugins)
	}

	childSet := map[string]bool{}
	for _, c := range codex.Children {
		childSet[c.Name] = true
	}
	if !childSet["sessions"] || !childSet["skills"] {
		t.Fatalf(".codex children = %v, want sessions and skills", childNames(codex))
	}

	if !resp.Summary.HasCodex {
		t.Fatal("summary.HasCodex = false, want true")
	}
	if resp.Summary.CodexSessions != 2 {
		t.Fatalf("summary.CodexSessions = %d, want 2", resp.Summary.CodexSessions)
	}
	if resp.Summary.CodexSkills != 1 {
		t.Fatalf("summary.CodexSkills = %d, want 1", resp.Summary.CodexSkills)
	}
}
```