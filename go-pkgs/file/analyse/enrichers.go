package analyse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

func semanticLinesForEntry(name, entryPath string) []SemanticLine {
	switch name {
	case ".codex":
		return enrichCodex(entryPath)
	case ".grok":
		return enrichGrok(entryPath)
	case ".cursor":
		return enrichCursor(entryPath)
	case ".knowledge-hub":
		return enrichKnowledgeHub(entryPath)
	case ".knowledge-index":
		return enrichKnowledgeIndex(entryPath)
	case ".openclaw":
		return enrichOpenclaw(entryPath)
	case ".opencode":
		return enrichOpencode(entryPath)
	default:
		return nil
	}
}

func enrichCodex(root string) []SemanticLine {
	sessionsDir := filepath.Join(root, "sessions")
	skillsDir := filepath.Join(root, "skills")
	rulesDir := filepath.Join(root, "rules")
	pluginsDir := filepath.Join(root, "plugins")
	cacheDir := filepath.Join(root, "cache")
	logsDir := filepath.Join(root, "logs")
	stateDir := filepath.Join(root, "state")
	memoriesDir := filepath.Join(root, "memories")
	historyFile := filepath.Join(root, "history.jsonl")
	modelsCache := filepath.Join(root, "models_cache.json")
	shellSnapshots := filepath.Join(root, "shell_snapshots")

	rollouts := countRolloutFiles(sessionsDir)
	skills := countTopLevelDirs(skillsDir)
	rules := countTopLevelFiles(rulesDir)
	plugins := countTopLevelDirs(pluginsDir)

	memoriesBytes := dirSizeIfExists(memoriesDir)
	for _, ent := range globBasenames(root, "memories_*.sqlite") {
		memoriesBytes += dirSizeIfExists(filepath.Join(root, ent))
	}

	return []SemanticLine{
		semanticCount("sessions", rollouts, "rollouts", dirSizeIfExists(sessionsDir)),
		semanticCount("skills", skills, skillUnit(skills), dirSizeIfExists(skillsDir)),
		semanticCount("rules", rules, ruleUnit(rules), dirSizeIfExists(rulesDir)),
		semanticCount("plugins", plugins, pluginUnit(plugins), dirSizeIfExists(pluginsDir)),
		semanticDash("cache", dirSizeIfExists(cacheDir)),
		semanticCount("logs", countSQLiteFamily(logsDir, "logs_"), "databases", sqliteFamilySize(logsDir, "logs_")),
		semanticCount("state", countSQLiteFamily(stateDir, "state_"), "databases", sqliteFamilySize(stateDir, "state_")),
		semanticCount("memories", countDirEntries(memoriesDir)+countSQLiteFamily(root, "memories_"), "items", memoriesBytes),
		semanticLines("history", historyLineCount(historyFile), "lines", fileSizeOrZero(historyFile)),
		semanticDash("models-cache", fileSizeOrZero(modelsCache)),
		semanticCount("shell-snapshots", countDirEntries(shellSnapshots), "entries", dirSizeIfExists(shellSnapshots)),
	}
}

func enrichGrok(root string) []SemanticLine {
	sessionsDir := filepath.Join(root, "sessions")
	projectsDir := filepath.Join(root, "projects")
	skillsDir := filepath.Join(root, "skills")
	downloadsDir := filepath.Join(root, "downloads")
	logsDir := filepath.Join(root, "logs")
	marketplaceDir := filepath.Join(root, "marketplace-cache")
	vendorDir := filepath.Join(root, "vendor")
	activeSessions := filepath.Join(root, "active_sessions.json")

	return []SemanticLine{
		semanticCount("sessions", countTopLevelDirs(sessionsDir), "sessions", dirSizeIfExists(sessionsDir)),
		semanticCount("projects", countTopLevelDirs(projectsDir), "projects", dirSizeIfExists(projectsDir)),
		semanticCount("skills", countTopLevelDirs(skillsDir), "skills", dirSizeIfExists(skillsDir)),
		semanticDash("downloads", dirSizeIfExists(downloadsDir)),
		semanticDash("logs", dirSizeIfExists(logsDir)),
		semanticCount("marketplace-cache", countTopLevelDirs(marketplaceDir), "entries", dirSizeIfExists(marketplaceDir)),
		semanticDash("vendor", dirSizeIfExists(vendorDir)),
		semanticCount("active-sessions", activeSessionCount(activeSessions), "sessions", fileSizeOrZero(activeSessions)),
	}
}

func enrichCursor(root string) []SemanticLine {
	projectsDir := filepath.Join(root, "projects")
	chatsDir := filepath.Join(root, "chats")
	skillsDir := filepath.Join(root, "skills-cursor")
	aiTracking := filepath.Join(root, "ai-tracking")

	return []SemanticLine{
		semanticCount("projects", countTopLevelDirs(projectsDir), "projects", dirSizeIfExists(projectsDir)),
		semanticCount("chats", countTopLevelDirs(chatsDir), "chats", dirSizeIfExists(chatsDir)),
		semanticCount("skills", countDirEntries(skillsDir), "items", dirSizeIfExists(skillsDir)),
		semanticDash("ai-tracking", dirSizeIfExists(aiTracking)),
	}
}

func enrichKnowledgeHub(root string) []SemanticLine {
	knowledges := filepath.Join(root, "knowledges")
	conversations := filepath.Join(root, "conversations")

	return []SemanticLine{
		semanticCount("knowledges", countKnowledgeEntries(knowledges), "entries", dirSizeIfExists(knowledges)),
		semanticCount("conversations", countDirEntries(conversations), "items", dirSizeIfExists(conversations)),
	}
}

func enrichKnowledgeIndex(root string) []SemanticLine {
	agents := filepath.Join(root, "agents")
	knowledgeBase := filepath.Join(root, "knowledge_base")
	conversations := filepath.Join(root, "conversations")

	return []SemanticLine{
		semanticCount("agents", countTopLevelDirs(agents), "agents", dirSizeIfExists(agents)),
		semanticCount("knowledge-base", countDirEntries(knowledgeBase), "entries", dirSizeIfExists(knowledgeBase)),
		semanticCount("conversations", countDirEntries(conversations), "items", dirSizeIfExists(conversations)),
	}
}

func enrichOpenclaw(root string) []SemanticLine {
	agents := filepath.Join(root, "agents")
	workspace := filepath.Join(root, "workspace")
	pluginSkills := filepath.Join(root, "plugin-skills")
	memory := filepath.Join(root, "memory")
	logs := filepath.Join(root, "logs")
	npm := filepath.Join(root, "npm")

	return []SemanticLine{
		semanticCount("agents", countTopLevelDirs(agents), "agents", dirSizeIfExists(agents)),
		semanticCount("workspace", countDirEntries(workspace), "items", dirSizeIfExists(workspace)),
		semanticCount("plugin-skills", countDirEntries(pluginSkills), "items", dirSizeIfExists(pluginSkills)),
		semanticDash("memory", dirSizeIfExists(memory)),
		semanticDash("logs", dirSizeIfExists(logs)),
		semanticDash("npm", dirSizeIfExists(npm)),
	}
}

func enrichOpencode(root string) []SemanticLine {
	binDir := filepath.Join(root, "bin")
	nodeModules := filepath.Join(root, "node_modules")

	return []SemanticLine{
		semanticCount("bin", countTopLevelFiles(binDir), "files", dirSizeIfExists(binDir)),
		semanticCount("node_modules", boolCount(dirExists(nodeModules)), "dir", dirSizeIfExists(nodeModules)),
	}
}

func semanticCount(key string, count int, unit string, bytes int64) SemanticLine {
	return SemanticLine{
		Key:       key,
		Count:     strconv.Itoa(count),
		Unit:      unit,
		Bytes:     bytes,
		SizeHuman: FormatSize(bytes),
	}
}

func semanticLines(key string, count int, unit string, bytes int64) SemanticLine {
	return semanticCount(key, count, unit, bytes)
}

func semanticDash(key string, bytes int64) SemanticLine {
	return SemanticLine{
		Key:       key,
		Count:     "—",
		Bytes:     bytes,
		SizeHuman: FormatSize(bytes),
	}
}

func skillUnit(n int) string {
	if n == 1 {
		return "skill"
	}
	return "skills"
}

func ruleUnit(n int) string {
	if n == 1 {
		return "rule"
	}
	return "rules"
}

func pluginUnit(n int) string {
	if n == 1 {
		return "plugin"
	}
	return "plugins"
}

func fileSizeOrZero(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func historyLineCount(path string) int {
	n, err := countTextLines(path)
	if err != nil {
		return 0
	}
	return n
}

func sqliteFamilySize(dir, prefix string) int64 {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var total int64
	for _, ent := range entries {
		name := ent.Name()
		if stringsHasPrefixSuffix(name, prefix, ".sqlite") ||
			stringsHasPrefixSuffix(name, prefix, ".sqlite-wal") ||
			stringsHasPrefixSuffix(name, prefix, ".sqlite-shm") {
			total += fileSizeOrZero(filepath.Join(dir, name))
		}
	}
	return total
}

func stringsHasPrefixSuffix(s, prefix, suffix string) bool {
	return len(s) >= len(prefix)+len(suffix) &&
		s[:len(prefix)] == prefix &&
		s[len(s)-len(suffix):] == suffix
}

func globBasenames(dir, pattern string) []string {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil
	}
	var names []string
	for _, m := range matches {
		names = append(names, filepath.Base(m))
	}
	return names
}

func countKnowledgeEntries(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, ent := range entries {
		if ent.Name() == ".git" {
			continue
		}
		n++
	}
	return n
}

func activeSessionCount(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return 0
	}
	switch v := parsed.(type) {
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		return 0
	}
}

func boolCount(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// TopicCounts extracts summary rollups from tool indicator dirs.
type TopicCounts struct {
	CodexSessions  int
	CodexSkills    int
	GrokSessions   int
	GrokProjects   int
	GrokSkills     int
	CursorProjects int
	CursorChats    int
	KHKnowledges   int
	KIAgents       int
	OpenclawAgents int
}

func topicCountsFromEntry(name, entryPath string) TopicCounts {
	var tc TopicCounts
	switch name {
	case ".codex":
		tc.CodexSessions = countRolloutFiles(filepath.Join(entryPath, "sessions"))
		tc.CodexSkills = countTopLevelDirs(filepath.Join(entryPath, "skills"))
	case ".grok":
		tc.GrokSessions = countTopLevelDirs(filepath.Join(entryPath, "sessions"))
		tc.GrokProjects = countTopLevelDirs(filepath.Join(entryPath, "projects"))
		tc.GrokSkills = countTopLevelDirs(filepath.Join(entryPath, "skills"))
	case ".cursor":
		tc.CursorProjects = countTopLevelDirs(filepath.Join(entryPath, "projects"))
		tc.CursorChats = countTopLevelDirs(filepath.Join(entryPath, "chats"))
	case ".knowledge-hub":
		tc.KHKnowledges = countKnowledgeEntries(filepath.Join(entryPath, "knowledges"))
	case ".knowledge-index":
		tc.KIAgents = countTopLevelDirs(filepath.Join(entryPath, "agents"))
	case ".openclaw":
		tc.OpenclawAgents = countTopLevelDirs(filepath.Join(entryPath, "agents"))
	}
	return tc
}