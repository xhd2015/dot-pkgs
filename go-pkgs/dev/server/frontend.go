package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// The generated wrapper preserves the project's config function and plugins.
// No import of vite is needed, so it can live outside frontend/node_modules.
func viteConfig(cfg Config, runDir, id string) (string, error) {
	v := cfg.Frontend
	configFile := v.ConfigFile
	if configFile == "" {
		configFile = "vite.config.ts"
	}
	configPath := filepath.Join(cfg.Root, v.Dir, configFile)
	pathJSON, _ := json.Marshal(filepath.ToSlash(configPath))
	idJSON, _ := json.Marshal(id)
	content := fmt.Sprintf(`import project from %s;
export default async (env) => {
  const base = await (typeof project === 'function' ? project(env) : project);
  return { ...base, plugins: [{
    name: 'dot-pkgs-dev-identity', enforce: 'pre',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (req.url?.split('?')[0] !== '/__dev/ready') return next();
        res.setHeader('Cache-Control', 'no-store');
        res.setHeader('Content-Type', 'text/plain');
        res.end(%s);
      });
    }
  }, ...(base.plugins || [])] };
};
`, pathJSON, idJSON)
	path := filepath.Join(runDir, "vite.config.mjs")
	return path, os.WriteFile(path, []byte(content), 0600)
}

func startVite(cfg Config, runDir, id string, port int) (*child, error) {
	path, err := viteConfig(cfg, runDir, id)
	if err != nil {
		return nil, err
	}
	args := append([]string{}, cfg.Frontend.Command...)
	args = append(args, "--config", path, "--host", "127.0.0.1", "--port", strconv.Itoa(port), "--strictPort", "--clearScreen", "false")
	return startChild(filepath.Join(cfg.Root, cfg.Frontend.Dir), args, nil, cfg)
}
