/*
Package sudosetup installs and detects persistent NOPASSWD sudoers rules for one
privileged command.

Callers supply Config (cache dir name, sudoers drop-in name, target username)
and Rule (command path plus optional args pattern). Manager writes
/etc/sudoers.d/<SudoersName> and ~/.cache/<CacheDirName>/sudo-setup-manifest.json
for persistent detection that does not rely on sudo timestamp cache alone.

Primary functions:
  - Detect: report installed state, cache warm, and non-interactive runnable
  - EnsureInstalled: one-time visudo-validated install (requires stdin TTY)
  - Remove: delete drop-in and manifest, flush sudo cache with sudo -k
  - IsInstalled: persistent drop-in + manifest match only

Tests inject fake FS and Runner; production callers use default osFS/execRunner.
*/
package sudosetup