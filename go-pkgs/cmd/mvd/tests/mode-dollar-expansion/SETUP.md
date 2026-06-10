## Steps
- Override Run to inject HOME and X environment variables for lls dollar-expansion support.
- HOME points to `req.WorkRoot/.lls-home` where the lls config with `"envs":["X"]` resides.
- X resolves to `req.WorkRoot/projects`.

```go
func runMvdDollar(t *testing.T, req *Request) (*Response, error) {
	bin := getMvdBin(t)
	cmd := exec.Command(bin, req.Args...)
	cmd.Env = append(os.Environ(),
		"MVD_DEBUG_CONFIG_HOME="+req.ConfigHome,
		"HOME="+filepath.Join(req.WorkRoot, ".lls-home"),
		"X="+filepath.Join(req.WorkRoot, "projects"),
	)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Output: string(out), ExitCode: exitCode}, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	_ = runMvdDollar
	bin := getMvdBin(t)
	cmd := exec.Command(bin, req.Args...)
	cmd.Env = append(os.Environ(),
		"MVD_DEBUG_CONFIG_HOME="+req.ConfigHome,
		"HOME="+filepath.Join(req.WorkRoot, ".lls-home"),
		"X="+filepath.Join(req.WorkRoot, "projects"),
	)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Output: string(out), ExitCode: exitCode}, nil
}
```
