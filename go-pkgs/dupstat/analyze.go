package dupstat

import (
	"fmt"
)

func Analyze(dir string, k int, threshold float64, algorithm string) ([]Group, error) {
	moduleRoot, err := FindModuleRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("find module root: %w", err)
	}

	modulePath, err := FindModulePath(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("read module path: %w", err)
	}

	files, err := FindGoFiles(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("find go files: %w", err)
	}

	var allFuncTokens []FunctionTokens
	for _, filePath := range files {
		pkgPath := PackagePath(moduleRoot, filePath)
		fullPkgPath := pkgPath
		if modulePath != "" && pkgPath != "" {
			fullPkgPath = modulePath + "/" + pkgPath
		} else if modulePath != "" {
			fullPkgPath = modulePath
		}

		funcs, err := ExtractFunctions(filePath, fullPkgPath)
		if err != nil {
			continue
		}
		for _, fn := range funcs {
			allFuncTokens = append(allFuncTokens, TokenizeFunction(fn))
		}
	}

	pairs := CompareFunctions(allFuncTokens, k, threshold, algorithm)
	groups := GroupPairs(pairs)

	return groups, nil
}
