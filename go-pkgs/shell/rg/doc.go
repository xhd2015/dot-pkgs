// Package rg discovers, installs, and invokes the BurntSushi ripgrep (`rg`) CLI.
//
// Discovery reuses lookpath + binaryversion (newest among candidates).
// Install downloads official GitHub release precompiled archives into
// ~/.local/bin when no rg is found. Unsupported GOOS/GOARCH pairs error out.
package rg
