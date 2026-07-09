package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	lessflags "github.com/xhd2015/less-flags"
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
  -h,--help                         show help message

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
	cfg, err := parseArgs(args, out)
	if err != nil {
		return err
	}
	if cfg.showHelp {
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

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	var originDomain *string
	var excludeOriginDomain *string
	var names []string
	var emails []string
	var notNames []string
	var notEmails []string

	remaining, err := lessflags.
		String("--origin-domain", &originDomain).
		String("--exclude-origin-domain", &excludeOriginDomain).
		StringSlice("--name", &names).
		StringSlice("--email", &emails).
		StringSlice("--not-name", &notNames).
		StringSlice("--not-email", &notEmails).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			cfg.showHelp = true
			return cfg, nil
		}
		return cfg, mapUnknownFlagErr(err)
	}
	if len(remaining) > 0 {
		return cfg, fmt.Errorf("unexpected arg: %s", remaining[0])
	}

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}

	// Preserve relative order within each flag family; process families in
	// declaration order matching common CLI usage.
	for _, raw := range names {
		check, err := parseFieldCheck("name", raw, false)
		if err != nil {
			return cfg, err
		}
		cfg.checks = append(cfg.checks, check)
	}
	for _, raw := range emails {
		check, err := parseFieldCheck("email", raw, false)
		if err != nil {
			return cfg, err
		}
		cfg.checks = append(cfg.checks, check)
	}
	for _, raw := range notNames {
		check, err := parseFieldCheck("name", raw, true)
		if err != nil {
			return cfg, err
		}
		cfg.checks = append(cfg.checks, check)
	}
	for _, raw := range notEmails {
		check, err := parseFieldCheck("email", raw, true)
		if err != nil {
			return cfg, err
		}
		cfg.checks = append(cfg.checks, check)
	}

	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func mapUnknownFlagErr(err error) error {
	const prefix = "unrecognized flag: "
	if msg := err.Error(); strings.HasPrefix(msg, prefix) {
		return fmt.Errorf("unknown flag: %s", strings.TrimPrefix(msg, prefix))
	}
	return err
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
