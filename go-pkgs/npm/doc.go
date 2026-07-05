/*
Package npm detects JavaScript package managers and builds install commands.

Detection inspects lockfiles, node_modules layout, and the packageManager field
in package.json. When multiple indicators are present, pnpm is preferred over
bun, then npm, then yarn.

Primary functions:
  - DetectFromDir: detect from a project directory
  - DetectFromNodeModules: detect from a node_modules absolute path
  - Resolve: pick a manager for running commands (auto or explicit)
  - InstallArgs: build install arguments for a manager
*/
package npm