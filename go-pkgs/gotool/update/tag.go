package update

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	errNoTag               = fmt.Errorf("no tag found")
	numericVersionPattern  = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
)

func getVersionPrefix(dir string) (string, error) {
	_, subPathList, err := getSubPath(dir)
	if err != nil {
		return "", err
	}
	if len(subPathList) == 0 {
		return "v", nil
	}
	return strings.Join(subPathList, "/") + "/v", nil
}

func addVersionPrefix(prefix string) string {
	if prefix == "" {
		return "v"
	}
	return strings.TrimSuffix(prefix, "/") + "/v"
}

func stripVersionPrefix(prefix, tag string) string {
	version := tag
	if strings.HasPrefix(tag, prefix) {
		version = tag[len(prefix):]
	}
	return "v" + strings.TrimPrefix(version, "v")
}

func getSubPath(dir string) (string, []string, error) {
	topLevel, err := showTopLevel(dir)
	if err != nil {
		return "", nil, err
	}
	topLevel = strings.TrimSpace(topLevel)

	absTopLevel, err := filepath.Abs(topLevel)
	if err != nil {
		return "", nil, err
	}
	absTopLevel, err = filepath.EvalSymlinks(absTopLevel)
	if err != nil {
		return "", nil, err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, err
	}
	absDir, err = filepath.EvalSymlinks(absDir)
	if err != nil {
		return "", nil, err
	}

	subPath, err := filepath.Rel(absTopLevel, absDir)
	if err != nil {
		return "", nil, err
	}
	var subPathList []string
	if subPath != "" && subPath != "." {
		subPathList = strings.Split(filepath.ToSlash(subPath), "/")
	}

	return topLevel, subPathList, nil
}

func getLatestVersionTag(dir, versionPrefix string) (string, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname", versionPrefix+"*")
	cmd.Dir = dir
	tagOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get tags for %s: %w", dir, err)
	}

	tags := strings.Split(strings.TrimSpace(string(tagOutput)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return "", fmt.Errorf("%w: %s", errNoTag, dir)
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		if versionPrefix != "" {
			if !strings.HasPrefix(tag, versionPrefix) {
				continue
			}
			basicName := strings.TrimPrefix(tag, versionPrefix)
			if strings.Contains(basicName, "/") {
				continue
			}
			if !numericVersionPattern.MatchString("v" + basicName) {
				continue
			}
		} else {
			if strings.Contains(tag, "/") {
				continue
			}
			if !numericVersionPattern.MatchString(tag) {
				continue
			}
		}

		return tag, nil
	}

	if versionPrefix != "" {
		return "", fmt.Errorf("%w: (%sv0.0.X) in %s", errNoTag, versionPrefix, dir)
	}
	return "", fmt.Errorf("%w: (v0.0.X) in %s", errNoTag, dir)
}

func showTopLevel(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}