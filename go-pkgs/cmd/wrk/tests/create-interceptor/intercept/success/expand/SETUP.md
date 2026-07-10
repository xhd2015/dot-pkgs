# Scenario

**Feature**: interceptor template expansion (`${name}`, `${name|shell_safe}`, array vars)

```
# vars + argv templates
intent_prompt = "/intent-route ${task}"
send = [ "wrk --no-interceptor ${args_shell_safe}", "agent-run … ${intent_prompt|shell_safe}" ]
argv includes --send ${send}
  -> fake receives expanded --send (possibly multiline)
```

## Steps

- Override config with the docs-style recipe (fake `kool` instead of real kool).
- Leaves set task text and assert shell_safe / multiline join.

```go
func Setup(t *testing.T, req *Request) error {
	// Replace simple config with recipe-style argv/vars.
	writeRecipeInterceptor(t, req.WrkHome)
	req.TaskFlag = "-t"
	return nil
}
```
