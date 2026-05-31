package dupstat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWordstatNoDupCoincidentalVars(t *testing.T) {
	pairs := analyzeTestDataDir(t, "coincidental-vars")
	for _, p := range pairs {
		t.Errorf("unexpected match: %s vs %s (j=%.2f c=%.2f o=%.2f)",
			p.FuncA.Name, p.FuncB.Name,
			p.WordJaccard, p.WordContainment, p.WordOverlap)
	}
}

func TestWordstatNoDupStructuralVocab(t *testing.T) {
	pairs := analyzeTestDataDir(t, "structural-vocab")
	for _, p := range pairs {
		t.Errorf("unexpected match: %s vs %s (j=%.2f c=%.2f o=%.2f)",
			p.FuncA.Name, p.FuncB.Name,
			p.WordJaccard, p.WordContainment, p.WordOverlap)
	}
}

func TestWordstatNoDupWrapperVsLarge(t *testing.T) {
	pairs := analyzeTestDataDir(t, "wrapper-vs-large")
	for _, p := range pairs {
		t.Errorf("unexpected match: %s vs %s (j=%.2f c=%.2f o=%.2f)",
			p.FuncA.Name, p.FuncB.Name,
			p.WordJaccard, p.WordContainment, p.WordOverlap)
	}
}

func TestWordstatNoDupStringVocabOverlap(t *testing.T) {
	pairs := analyzeTestDataDir(t, "string-vocab-overlap")
	for _, p := range pairs {
		t.Errorf("unexpected match: %s vs %s (j=%.2f c=%.2f o=%.2f)",
			p.FuncA.Name, p.FuncB.Name,
			p.WordJaccard, p.WordContainment, p.WordOverlap)
	}
}

func TestWordstatDupPreserved(t *testing.T) {
	pairs := analyzeTestDataDirAt(t, filepath.Join("..", "cmd", "code-dup-stat", "testdata", "wordstat-dup"))
	if len(pairs) == 0 {
		t.Error("expected ProcessUser/HandleRequest to be detected")
		return
	}
	found := false
	for _, p := range pairs {
		if (p.FuncA.Name == "ProcessUser" && p.FuncB.Name == "HandleRequest") ||
			(p.FuncA.Name == "HandleRequest" && p.FuncB.Name == "ProcessUser") {
			found = true
			break
		}
	}
	if !found {
		t.Error("ProcessUser/HandleRequest pair not found")
	}
}

func analyzeTestDataDir(t *testing.T, dirName string) []FuncPair {
	baseDir := filepath.Join("..", "cmd", "code-dup-stat", "testdata", "wordstat-no-dup", dirName)
	return analyzeTestDataDirAt(t, baseDir)
}

func analyzeTestDataDirAt(t *testing.T, baseDir string) []FuncPair {
	t.Helper()

	var allFuncTokens []FunctionTokens
	collectGoFiles(t, baseDir, &allFuncTokens)

	if len(allFuncTokens) < 2 {
		t.Fatalf("expected at least 2 functions from %s, got %d", baseDir, len(allFuncTokens))
	}

	pairs := CompareFunctions(allFuncTokens, 5, 0.5, AlgoWordstat)
	return pairs
}

func collectGoFiles(t *testing.T, dir string, tokens *[]FunctionTokens) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Logf("read dir %s: %v", dir, err)
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			collectGoFiles(t, fullPath, tokens)
		} else if filepath.Ext(entry.Name()) == ".go" {
			funcs, err := ExtractFunctions(fullPath, "testpkg")
			if err != nil {
				t.Logf("extract %s: %v", fullPath, err)
				continue
			}
			for _, fn := range funcs {
				*tokens = append(*tokens, TokenizeFunction(fn))
			}
		}
	}
}
