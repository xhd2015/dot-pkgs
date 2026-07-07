package analyse

import (
	"fmt"
	"strings"
)

// FormatEntryBlock renders one entry as a human-readable block.
func FormatEntryBlock(entry EntryResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "> %s\n", entry.Name)

	if entry.Kind == EntryKindFile {
		fmt.Fprintf(&b, "  size    %s\n", entry.SizeHuman)
		fmt.Fprintf(&b, "  lines   %s\n", entry.Lines)
		return strings.TrimRight(b.String(), "\n")
	}

	for _, child := range entry.Children {
		fmt.Fprintf(&b, "  > %-32s %s\n", child.Name, child.SizeHuman)
	}

	if len(entry.Semantic) > 0 && len(entry.Children) > 0 {
		b.WriteString("\n")
	}
	for _, line := range entry.Semantic {
		b.WriteString(formatSemanticLine(line))
		b.WriteString("\n")
	}

	if entry.Aggregates.GitRepos > 0 {
		fmt.Fprintf(&b, "  git-dirs          %d\n", entry.Aggregates.GitRepos)
	}
	if entry.Aggregates.LinkedWorktrees > 0 {
		fmt.Fprintf(&b, "  worktrees         %d linked\n", entry.Aggregates.LinkedWorktrees)
	}
	if entry.Aggregates.NodeModulesDirs > 0 {
		fmt.Fprintf(&b, "  node_modules      %d dirs\n", entry.Aggregates.NodeModulesDirs)
	}

	return strings.TrimRight(b.String(), "\n")
}

func formatSemanticLine(line SemanticLine) string {
	countPart := line.Count
	if line.Unit != "" && line.Count != "—" {
		countPart = fmt.Sprintf("%s %s", line.Count, line.Unit)
	} else if line.Count == "—" {
		countPart = "—"
	}
	return fmt.Sprintf("  %-18s %-16s %s", line.Key, countPart, line.SizeHuman)
}

// FormatSummaryLines renders the analyse-files summary block.
func FormatSummaryLines(summary ScanSummary) []string {
	lines := []string{
		"analyse-files summary",
		fmt.Sprintf("  home:              %s", summary.Home),
		fmt.Sprintf("  entries scanned:   %d   (%d dirs, %d files)", summary.EntryCount, summary.DirCount, summary.FileCount),
		fmt.Sprintf("  total size:        %s", summary.TotalHuman),
		"",
		fmt.Sprintf("  git repos:         %d", summary.GitRepos),
		fmt.Sprintf("  linked worktrees:  %d", summary.LinkedWorktrees),
	}

	if summary.HasCodex {
		lines = append(lines, "",
			fmt.Sprintf("  codex sessions:    %d", summary.CodexSessions),
			fmt.Sprintf("  codex skills:      %d", summary.CodexSkills),
		)
	}
	if summary.HasGrok {
		lines = append(lines, "",
			fmt.Sprintf("  grok sessions:     %d", summary.GrokSessions),
			fmt.Sprintf("  grok projects:     %d", summary.GrokProjects),
			fmt.Sprintf("  grok skills:       %d", summary.GrokSkills),
		)
	}
	if summary.HasCursor {
		lines = append(lines, "",
			fmt.Sprintf("  cursor projects:   %d", summary.CursorProjects),
			fmt.Sprintf("  cursor chats:      %d", summary.CursorChats),
		)
	}
	if summary.HasKnowledgeHub {
		lines = append(lines, "",
			fmt.Sprintf("  knowledge-hub knowledges: %d", summary.KHKnowledges),
		)
	}
	if summary.HasKnowledgeIndex {
		lines = append(lines, "",
			fmt.Sprintf("  knowledge-index agents: %d", summary.KIAgents),
		)
	}
	if summary.HasOpenclaw {
		lines = append(lines, "",
			fmt.Sprintf("  openclaw agents:   %d", summary.OpenclawAgents),
		)
	}

	lines = append(lines, "",
		fmt.Sprintf("  node_modules:      %d dirs", summary.NodeModulesDirs),
		"",
		"  largest entries:   (top 5)",
	)
	for _, largest := range summary.Largest {
		lines = append(lines, fmt.Sprintf("    %-18s %s", largest.Name, largest.SizeHuman))
	}
	return lines
}

// SummaryToDone builds the structured done frame payload for streaming callers.
func SummaryToDone(summary ScanSummary, entries []EntryResult) map[string]any {
	done := map[string]any{
		"home":              summary.Home,
		"entries":           summary.EntryCount,
		"dirs":              summary.DirCount,
		"files":             summary.FileCount,
		"total_bytes":       summary.TotalBytes,
		"total_human":       summary.TotalHuman,
		"git_repos":         summary.GitRepos,
		"linked_worktrees":  summary.LinkedWorktrees,
		"node_modules_dirs": summary.NodeModulesDirs,
		"entry_results":     entries,
		"largest_entries":   summary.Largest,
	}
	if summary.HasCodex {
		done["codex_sessions"] = summary.CodexSessions
		done["codex_skills"] = summary.CodexSkills
	}
	if summary.HasGrok {
		done["grok_sessions"] = summary.GrokSessions
		done["grok_projects"] = summary.GrokProjects
		done["grok_skills"] = summary.GrokSkills
	}
	if summary.HasCursor {
		done["cursor_projects"] = summary.CursorProjects
		done["cursor_chats"] = summary.CursorChats
	}
	if summary.HasKnowledgeHub {
		done["knowledge_hub_knowledges"] = summary.KHKnowledges
	}
	if summary.HasKnowledgeIndex {
		done["knowledge_index_agents"] = summary.KIAgents
	}
	if summary.HasOpenclaw {
		done["openclaw_agents"] = summary.OpenclawAgents
	}
	return done
}