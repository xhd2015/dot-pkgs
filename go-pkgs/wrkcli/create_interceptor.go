package wrkcli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// Reserved builtin template variable names (user vars cannot override).
var interceptorBuiltinNames = map[string]struct{}{
	"work_dir":        {},
	"orig_wd":         {},
	"source":          {},
	"spawn_target":    {},
	"task":            {},
	"task_slug":       {},
	"args":            {},
	"args_shell_safe": {},
	"wrk_home":        {},
}

// createInterceptorInput carries runtime values for template expansion.
type createInterceptorInput struct {
	wrkHome     string
	workDir     string
	origWd      string
	source      string
	spawnTarget string
	task        string
	args        []string // original CLI args into Run (after binary name)
}

// runCreateInterceptor expands config argv/vars and execs the interceptor,
// inheriting stdio and env. Child non-zero exit becomes ExitCodeError.
func runCreateInterceptor(ic *CreateInterceptor, in createInterceptorInput) error {
	if ic == nil || !ic.Enabled {
		return fmt.Errorf("wrk: create interceptor is not enabled")
	}
	if len(ic.Argv) == 0 {
		return fmt.Errorf("wrk: create interceptor enabled but argv is empty")
	}

	expanded, err := expandCreateInterceptor(ic, in)
	if err != nil {
		return err
	}
	if len(expanded) == 0 {
		return fmt.Errorf("wrk: create interceptor argv expanded to empty")
	}

	cmd := exec.Command(expanded[0], expanded[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 0 {
				code = 1
			}
			return ExitCodeError{Code: code}
		}
		return fmt.Errorf("wrk: create interceptor exec: %w", err)
	}
	return nil
}

// expandCreateInterceptor returns fully expanded argv tokens.
func expandCreateInterceptor(ic *CreateInterceptor, in createInterceptorInput) ([]string, error) {
	values, err := seedInterceptorBuiltins(in)
	if err != nil {
		return nil, err
	}
	if err := expandInterceptorUserVars(ic.Vars, values); err != nil {
		return nil, err
	}

	out := make([]string, len(ic.Argv))
	for i, tmpl := range ic.Argv {
		s, err := expandInterceptorTemplate(tmpl, values)
		if err != nil {
			return nil, err
		}
		out[i] = s
	}
	return out, nil
}

func seedInterceptorBuiltins(in createInterceptorInput) (map[string]string, error) {
	taskSlug := ""
	if strings.TrimSpace(in.task) != "" {
		taskSlug = slugify(in.task)
		if taskSlug == "" {
			return nil, fmt.Errorf("wrk: task description %q produces an empty slug", in.task)
		}
	}

	argsShellParts := make([]string, len(in.args))
	for i, a := range in.args {
		argsShellParts[i] = ShellSafeQuote(a)
	}

	return map[string]string{
		"work_dir":        in.workDir,
		"orig_wd":         in.origWd,
		"source":          in.source,
		"spawn_target":    in.spawnTarget,
		"task":            in.task,
		"task_slug":       taskSlug,
		"args":            strings.Join(in.args, " "),
		"args_shell_safe": strings.Join(argsShellParts, " "),
		"wrk_home":        in.wrkHome,
	}, nil
}

func expandInterceptorUserVars(vars map[string]VarValue, values map[string]string) error {
	if len(vars) == 0 {
		return nil
	}

	// Reject builtin overrides and collect dependency edges among user vars.
	names := make([]string, 0, len(vars))
	deps := make(map[string][]string, len(vars))
	for name, vv := range vars {
		if !isInterceptorVarName(name) {
			return fmt.Errorf("wrk: invalid interceptor var name %q", name)
		}
		if _, reserved := interceptorBuiltinNames[name]; reserved {
			return fmt.Errorf("wrk: interceptor var %q overrides reserved builtin", name)
		}
		names = append(names, name)
		refs, err := interceptorVarRefs(vv)
		if err != nil {
			return err
		}
		var userDeps []string
		for _, ref := range refs {
			if _, isUser := vars[ref]; isUser {
				userDeps = append(userDeps, ref)
			}
		}
		deps[name] = userDeps
	}

	order, err := topoSortInterceptorVars(names, deps)
	if err != nil {
		return err
	}

	for _, name := range order {
		vv := vars[name]
		expanded, err := expandInterceptorVarValue(vv, values)
		if err != nil {
			return err
		}
		values[name] = expanded
	}
	return nil
}

func expandInterceptorVarValue(vv VarValue, values map[string]string) (string, error) {
	if vv.IsList {
		parts := make([]string, len(vv.Lines))
		for i, line := range vv.Lines {
			s, err := expandInterceptorTemplate(line, values)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return strings.Join(parts, "\n"), nil
	}
	return expandInterceptorTemplate(vv.Scalar, values)
}

func interceptorVarRefs(vv VarValue) ([]string, error) {
	var refs []string
	seen := map[string]struct{}{}
	add := func(tmpl string) error {
		names, err := collectInterceptorRefs(tmpl)
		if err != nil {
			return err
		}
		for _, n := range names {
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			refs = append(refs, n)
		}
		return nil
	}
	if vv.IsList {
		for _, line := range vv.Lines {
			if err := add(line); err != nil {
				return nil, err
			}
		}
		return refs, nil
	}
	if err := add(vv.Scalar); err != nil {
		return nil, err
	}
	return refs, nil
}

func collectInterceptorRefs(tmpl string) ([]string, error) {
	var refs []string
	i := 0
	for i < len(tmpl) {
		if tmpl[i] != '$' || i+1 >= len(tmpl) || tmpl[i+1] != '{' {
			i++
			continue
		}
		close := strings.IndexByte(tmpl[i+2:], '}')
		if close < 0 {
			return nil, fmt.Errorf("wrk: unclosed interceptor template in %q", tmpl)
		}
		body := tmpl[i+2 : i+2+close]
		name, _, err := parseInterceptorTemplateBody(body)
		if err != nil {
			return nil, err
		}
		refs = append(refs, name)
		i = i + 2 + close + 1
	}
	return refs, nil
}

func topoSortInterceptorVars(names []string, deps map[string][]string) ([]string, error) {
	// Kahn's algorithm; stable-ish by iterating names as given (map order).
	inDegree := make(map[string]int, len(names))
	// reverse edges: dep -> dependents
	rev := make(map[string][]string, len(names))
	for _, n := range names {
		inDegree[n] = 0
	}
	for _, n := range names {
		for _, d := range deps[n] {
			if _, ok := inDegree[d]; !ok {
				// dependency on unknown user var should not happen (only user deps collected)
				continue
			}
			inDegree[n]++
			rev[d] = append(rev[d], n)
		}
	}
	var queue []string
	for _, n := range names {
		if inDegree[n] == 0 {
			queue = append(queue, n)
		}
	}
	var order []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		for _, m := range rev[n] {
			inDegree[m]--
			if inDegree[m] == 0 {
				queue = append(queue, m)
			}
		}
	}
	if len(order) != len(names) {
		return nil, fmt.Errorf("wrk: cycle in create interceptor vars")
	}
	return order, nil
}

// expandInterceptorTemplate substitutes ${name} and ${name|shell_safe}.
func expandInterceptorTemplate(tmpl string, values map[string]string) (string, error) {
	var b strings.Builder
	i := 0
	for i < len(tmpl) {
		if tmpl[i] != '$' || i+1 >= len(tmpl) || tmpl[i+1] != '{' {
			b.WriteByte(tmpl[i])
			i++
			continue
		}
		close := strings.IndexByte(tmpl[i+2:], '}')
		if close < 0 {
			return "", fmt.Errorf("wrk: unclosed interceptor template in %q", tmpl)
		}
		body := tmpl[i+2 : i+2+close]
		name, filter, err := parseInterceptorTemplateBody(body)
		if err != nil {
			return "", err
		}
		val, ok := values[name]
		if !ok {
			return "", fmt.Errorf("wrk: unknown interceptor template variable %q", name)
		}
		switch filter {
		case "":
			b.WriteString(val)
		case "shell_safe":
			b.WriteString(ShellSafeQuote(val))
		default:
			return "", fmt.Errorf("wrk: unknown interceptor template filter %q", filter)
		}
		i = i + 2 + close + 1
	}
	return b.String(), nil
}

func parseInterceptorTemplateBody(body string) (name, filter string, err error) {
	if body == "" {
		return "", "", fmt.Errorf("wrk: empty interceptor template variable")
	}
	if strings.Contains(body, "|") {
		parts := strings.SplitN(body, "|", 2)
		name = parts[0]
		filter = parts[1]
	} else {
		name = body
	}
	if !isInterceptorVarName(name) {
		return "", "", fmt.Errorf("wrk: invalid interceptor template variable %q", name)
	}
	if filter != "" && !isInterceptorVarName(filter) {
		return "", "", fmt.Errorf("wrk: unknown interceptor template filter %q", filter)
	}
	return name, filter, nil
}

func isInterceptorVarName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
