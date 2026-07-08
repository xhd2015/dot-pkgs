package wrkcli

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/xhd2015/skills/install"
	"github.com/xhd2015/skills/skill_file"
	lessflags "github.com/xhd2015/less-flags"
)

//go:embed SKILL.md
var skillContent string

const skillSubcommandName = "wrk"

var wrkModeFlags = map[string]struct{}{
	"--done":                {},
	"--merge-back":          {},
	"-l":                    {},
	"--list":                {},
	"--status":              {},
	"--repos":               {},
	"--projects":            {},
	"--fetch":               {},
	"--color":               {},
	"--add":                 {},
	"--rm":                  {},
	"--confirm-from-stdin":  {},
	"-y":                    {},
	"--yes":                 {},
	"--no-in-module-replace": {},
	"--all-deps":            {},
	"--dep":                 {},
	"-t":                    {},
	"--task":                {},
	"--set-task":            {},
	"--where":               {},
}

func runSkill(origWd string, args []string, wrkHome string) error {
	if flag, found := findWrkModeFlag(args); found {
		return fmt.Errorf("wrk: %s is mutually exclusive with skill subcommand", flag)
	}
	if len(args) == 0 {
		return fmt.Errorf("wrk: skill requires a subcommand (list, show, install)")
	}
	subcmd := args[0]
	subArgs := args[1:]
	switch subcmd {
	case "list":
		return runSkillList(subArgs)
	case "show":
		return runSkillShow(subArgs)
	case "install":
		return runSkillInstall(subArgs)
	default:
		return fmt.Errorf("wrk: unknown skill subcommand %q", subcmd)
	}
}

func findWrkModeFlag(args []string) (string, bool) {
	skipValue := false
	for _, arg := range args {
		if skipValue {
			skipValue = false
			continue
		}
		if _, ok := wrkModeFlags[arg]; ok {
			return arg, true
		}
		if strings.HasPrefix(arg, "-") {
			if _, ok := flagValueArgs[arg]; ok {
				skipValue = true
			}
		}
	}
	return "", false
}

func runSkillList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	fmt.Println(skillSubcommandName)
	return nil
}

func runSkillShow(args []string) error {
	headerOnly, err := parseSkillShowArgs(args)
	if err != nil {
		return err
	}
	content := skillContent
	if headerOnly {
		out, err := skill_file.FormatHeaderWithDelimiters(content)
		if err != nil {
			return fmt.Errorf("wrk: parse skill header: %w", err)
		}
		fmt.Print(out)
		return nil
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	fmt.Print(content)
	return nil
}

func parseSkillShowArgs(args []string) (headerOnly bool, err error) {
	var header bool
	remaining, err := lessflags.Bool("--header", &header).Parse(args)
	if err != nil {
		return false, err
	}
	if len(remaining) > 0 {
		return false, fmt.Errorf("wrk: unknown option %s", remaining[0])
	}
	return header, nil
}

func runSkillInstall(args []string) error {
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: skillSubcommandName,
		SkillContent: skillContent,
		Usage:        "wrk skill install",
	}, args)
}