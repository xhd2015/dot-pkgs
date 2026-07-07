package analyse

import (
	"regexp"
	"strings"
	"testing"
)

func TestFormatEntryBlockCodexSemanticOrder(t *testing.T) {
	entry := EntryResult{
		Name: ".codex",
		Kind: EntryKindDir,
		Children: []ChildLine{
			{Name: "sessions", SizeHuman: "12 KB"},
			{Name: "skills", SizeHuman: "4 KB"},
		},
		Semantic: []SemanticLine{
			{Key: "sessions", Count: "2", Unit: "rollouts", SizeHuman: "12 KB"},
			{Key: "skills", Count: "1", Unit: "skill", SizeHuman: "4 KB"},
		},
	}
	block := FormatEntryBlock(entry)
	childIdx := strings.Index(block, "> sessions")
	semanticIdx := strings.Index(block, "2 rollouts")
	if childIdx < 0 || semanticIdx < 0 || childIdx > semanticIdx {
		t.Fatalf("child lines should precede semantic lines:\n%s", block)
	}
	if !strings.Contains(block, "1 skill") {
		t.Fatalf("missing skill semantic line:\n%s", block)
	}
}

func TestFormatEntryBlockGitDirs(t *testing.T) {
	entry := EntryResult{
		Name: ".",
		Kind: EntryKindDir,
		Aggregates: Aggregates{GitRepos: 1},
	}
	block := FormatEntryBlock(entry)
	re := regexp.MustCompile(`(?m)^\s*git-dirs\s+1\s*$`)
	if !re.MatchString(block) {
		t.Fatalf("expected git-dirs 1, got:\n%s", block)
	}
}

func TestFormatEntryBlockNodeModulesAggregate(t *testing.T) {
	entry := EntryResult{
		Name: "nm-entry",
		Kind: EntryKindDir,
		Children: []ChildLine{{Name: "node_modules", SizeHuman: "1 KB"}},
		Aggregates: Aggregates{NodeModulesDirs: 2},
	}
	block := FormatEntryBlock(entry)
	if !strings.Contains(block, "> node_modules") {
		t.Fatalf("missing child node_modules:\n%s", block)
	}
	re := regexp.MustCompile(`(?m)^\s*node_modules\s+2\s+dirs\s*$`)
	if !re.MatchString(block) {
		t.Fatalf("missing node_modules 2 dirs aggregate:\n%s", block)
	}
}

func TestFormatSummaryTopicPresent(t *testing.T) {
	withCodex := FormatSummaryLines(ScanSummary{
		Home: "/tmp/home", EntryCount: 1, DirCount: 1, HasCodex: true,
		CodexSessions: 2, CodexSkills: 1,
	})
	joined := strings.Join(withCodex, "\n")
	if !strings.Contains(joined, "codex sessions:") || !strings.Contains(joined, "codex skills:") {
		t.Fatalf("expected codex summary lines:\n%s", joined)
	}

	withoutGrok := FormatSummaryLines(ScanSummary{Home: "/tmp/home", HasGrok: false})
	joined = strings.Join(withoutGrok, "\n")
	for _, needle := range []string{"grok sessions:", "grok projects:", "grok skills:"} {
		if strings.Contains(joined, needle) {
			t.Fatalf("unexpected %q in summary without .grok:\n%s", needle, joined)
		}
	}
}