package dotenv

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// Init loads environment variables from .env and .env.local files.
// .env.local overrides .env values.
// Missing files are silently ignored, but other errors are returned.
func Init() error {
	vars, err := loadFiles([]string{".env", ".env.local"}, true)
	if err != nil {
		return err
	}
	applyEnvVars(vars)
	return nil
}

// LoadFile loads environment variables from a single file.
// Returns error if the file cannot be read (including if it doesn't exist).
func LoadFile(filename string) error {
	vars, err := loadFiles([]string{filename}, false)
	if err != nil {
		return err
	}
	applyEnvVars(vars)
	return nil
}

// LoadFiles loads environment variables from multiple files.
// Later files override earlier ones.
// Missing files are silently ignored, but other errors are returned.
func LoadFiles(filenames ...string) error {
	vars, err := loadFiles(filenames, true)
	if err != nil {
		return err
	}
	applyEnvVars(vars)
	return nil
}

func loadFiles(files []string, allowMissing bool) (map[string]string, error) {
	envVars := make(map[string]string)

	for _, filename := range files {
		vars, err := readFile(filename)
		if err != nil {
			if allowMissing && errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for k, v := range vars {
			envVars[k] = v
		}
	}

	return envVars, nil
}

func applyEnvVars(vars map[string]string) {
	for k, v := range vars {
		os.Setenv(k, v)
	}
}

func readFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = trimQuotes(value)

		vars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return vars, nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
