// Package service is the shared service-management core used by hosts that run
// long-lived background work: eluc's system services and ai-critic's managed
// services.
//
// It owns the mechanics both hosts need, so the behaviour cannot drift:
//
//   - Definition: one managed command, its working directory and env layers.
//   - Supervisor: start, stop, restart, crash backoff, log files, runtime
//     pid/pgid records and orphan reclaim for a set of definitions.
//   - Process primitives: liveness (zombie-aware), process groups, port
//     owners, and the command line used to detect pid recycling.
//   - System services: the interfaces an in-process subsystem ("embedded"
//     service) implements so a host can list, start, stop and restart it next
//     to child-process services.
//
// The package deliberately contains no HTTP handlers and no storage policy:
// hosts decide where definitions live, how status is rendered, and which
// transport exposes it. See Supervisor's Options for the paths it does own
// (runtime state and logs).
package service
