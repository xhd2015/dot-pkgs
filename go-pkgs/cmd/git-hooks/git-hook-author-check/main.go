package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
)

const help = `
Usage: git-hook-author-check [OPTIONS]

Check the effective Git author name and email for the commit.

Options:
  --name CONDITION                  require author name to match CONDITION
  --email CONDITION                 require author email to match CONDITION
  --not-name CONDITION              require author name not to match CONDITION
  --not-email CONDITION             require author email not to match CONDITION
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message

Conditions:
  value                             exact full match, case-insensitive
  eq:value                          exact full match, case-insensitive
  contains:value                    substring match, case-insensitive
  starts-with:value                 prefix match, case-insensitive
  ends-with:value                   suffix match, case-insensitive
  !condition                        negate a condition
`

var errAuthorCheckFailed = errors.New("author check failed")

type config struct {
	domainFilter githook.DomainFilter
	checks       []fieldCheck
	showHelp     bool
}

type fieldCheck struct {
	field     string
	condition condition
}

type condition struct {
	op      conditionOp
	value   string
	negated bool
	raw     string
}

type conditionOp string

const (
	conditionEqual      conditionOp = "equals"
	conditionContains   conditionOp = "contains"
	conditionStartsWith conditionOp = "starts with"
	conditionEndsWith   conditionOp = "ends with"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errAuthorCheckFailed) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-author-check: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}
	if cfg.showHelp {
		fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		return nil
	}
	if len(cfg.checks) == 0 {
		return fmt.Errorf("requires at least one of --name, --email, --not-name, --not-email")
	}

	shouldRun, err := cfg.domainFilter.ShouldRun()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	a, err := githook.EffectiveAuthor()
	if err != nil {
		return err
	}

	var failures []string
	for _, check := range cfg.checks {
		actual := a.Name
		if check.field == "email" {
			actual = a.Email
		}
		if !check.condition.matches(actual) {
			failures = append(failures, fmt.Sprintf("%s %q does not satisfy %s", check.field, actual, check.condition.describe()))
		}
	}
	if len(failures) == 0 {
		return nil
	}

	fmt.Fprintln(out, "git author check failed:")
	for _, failure := range failures {
		fmt.Fprintf(out, "  %s\n", failure)
	}
	return errAuthorCheckFailed
}

func parseArgs(args []string) (config, error) {
	var cfg config
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if matched, next, err := githook.ParseDomainFlag(args, i, &cfg.domainFilter); matched {
			if err != nil {
				return cfg, err
			}
			i = next
			continue
		}
		switch {
		case arg == "-h" || arg == "--help":
			cfg.showHelp = true
			return cfg, nil
		case arg == "--name":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--name requires a value")
			}
			check, err := parseFieldCheck("name", args[i], false)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case strings.HasPrefix(arg, "--name="):
			check, err := parseFieldCheck("name", strings.TrimPrefix(arg, "--name="), false)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case arg == "--email":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--email requires a value")
			}
			check, err := parseFieldCheck("email", args[i], false)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case strings.HasPrefix(arg, "--email="):
			check, err := parseFieldCheck("email", strings.TrimPrefix(arg, "--email="), false)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case arg == "--not-name":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--not-name requires a value")
			}
			check, err := parseFieldCheck("name", args[i], true)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case strings.HasPrefix(arg, "--not-name="):
			check, err := parseFieldCheck("name", strings.TrimPrefix(arg, "--not-name="), true)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case arg == "--not-email":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--not-email requires a value")
			}
			check, err := parseFieldCheck("email", args[i], true)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case strings.HasPrefix(arg, "--not-email="):
			check, err := parseFieldCheck("email", strings.TrimPrefix(arg, "--not-email="), true)
			if err != nil {
				return cfg, err
			}
			cfg.checks = append(cfg.checks, check)
		case strings.HasPrefix(arg, "-"):
			return cfg, fmt.Errorf("unknown flag: %s", arg)
		default:
			return cfg, fmt.Errorf("unexpected arg: %s", arg)
		}
	}
	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parseFieldCheck(field string, raw string, negate bool) (fieldCheck, error) {
	cond, err := parseCondition(raw)
	if err != nil {
		return fieldCheck{}, err
	}
	if negate {
		cond.negated = !cond.negated
	}
	return fieldCheck{
		field:     field,
		condition: cond,
	}, nil
}

func parseCondition(raw string) (condition, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return condition{}, fmt.Errorf("condition must not be empty")
	}

	negated := false
	for {
		lower := strings.ToLower(s)
		switch {
		case strings.HasPrefix(s, "!"):
			negated = !negated
			s = strings.TrimSpace(strings.TrimPrefix(s, "!"))
		case strings.HasPrefix(lower, "not:"):
			negated = !negated
			s = strings.TrimSpace(s[len("not:"):])
		default:
			if s == "" {
				return condition{}, fmt.Errorf("condition must not be empty")
			}
			return parsePositiveCondition(raw, s, negated)
		}
	}
}

func parsePositiveCondition(raw string, s string, negated bool) (condition, error) {
	type prefix struct {
		text string
		op   conditionOp
	}
	prefixes := []prefix{
		{text: "starts-with:", op: conditionStartsWith},
		{text: "starts with:", op: conditionStartsWith},
		{text: "startswith:", op: conditionStartsWith},
		{text: "prefix:", op: conditionStartsWith},
		{text: "ends-with:", op: conditionEndsWith},
		{text: "ends with:", op: conditionEndsWith},
		{text: "endswith:", op: conditionEndsWith},
		{text: "suffix:", op: conditionEndsWith},
		{text: "contains:", op: conditionContains},
		{text: "equals:", op: conditionEqual},
		{text: "eq:", op: conditionEqual},
	}
	lower := strings.ToLower(s)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix.text) {
			value := s[len(prefix.text):]
			if value == "" {
				return condition{}, fmt.Errorf("condition value must not be empty: %s", raw)
			}
			return condition{op: prefix.op, value: value, negated: negated, raw: raw}, nil
		}
	}
	return condition{op: conditionEqual, value: s, negated: negated, raw: raw}, nil
}

func (c condition) matches(actual string) bool {
	actual = strings.ToLower(actual)
	value := strings.ToLower(c.value)
	var matched bool
	switch c.op {
	case conditionEqual:
		matched = actual == value
	case conditionContains:
		matched = strings.Contains(actual, value)
	case conditionStartsWith:
		matched = strings.HasPrefix(actual, value)
	case conditionEndsWith:
		matched = strings.HasSuffix(actual, value)
	default:
		matched = false
	}
	if c.negated {
		return !matched
	}
	return matched
}

func (c condition) describe() string {
	prefix := ""
	if c.negated {
		prefix = "not "
	}
	return fmt.Sprintf("%s%s %q", prefix, c.op, c.value)
}
