/*
Package logs provides utilities for real-time log file monitoring and output formatting.

The package offers functionality to efficiently watch log files for changes and
output new content with customizable prefixes. It uses fsnotify for efficient
file system event monitoring instead of polling.

Key features:
  - Real-time monitoring of log files using fsnotify
  - Outputs only new content when files are modified
  - Handles file creation, modification, and truncation
  - Support for callback functions to process new content
  - Prefixes each output chunk with customizable text
  - Context-based cancellation for graceful shutdown

Primary functions:
  - Watch: Monitor a file and process new content with a callback function
  - Pipe: Monitor any file with a custom prefix (uses Watch internally)

Example usage with callback:

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logs.Watch(ctx, "/path/to/app.log", func(content []byte) error {
	    // Process the content however you want
	    fmt.Print(string(content))
	    return nil
	})
	if err != nil {
	    log.Fatalf("Watcher stopped: %v", err)
	}

Example usage with prefix:

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logs.Pipe(ctx, "/path/to/app.log", "app: ", os.Stdout)
	if err != nil {
	    log.Fatalf("Error watching log: %v", err)
	}
*/
package logs
