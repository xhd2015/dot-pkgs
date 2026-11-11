package service

import "time"

// Crash-restart pacing. A service that exits right after starting is retried
// with exponential backoff; one that stayed up long enough to be considered
// stable gets a fresh budget.
const (
	// RestartBackoffBase is the first delay after one failure.
	RestartBackoffBase = time.Second
	// RestartBackoffMax caps the delay so a broken service is retried at least
	// every five minutes.
	RestartBackoffMax = 5 * time.Minute
	// StableRunResetWindow is how long a process must stay up for its exit to
	// stop counting as a crash loop.
	StableRunResetWindow = 30 * time.Second
)

// BackoffDelay returns how long to wait after the given number of consecutive
// failures: 1s, 2s, 4s, … capped at RestartBackoffMax. Zero or negative means
// no wait.
func BackoffDelay(failures int) time.Duration {
	if failures <= 0 {
		return 0
	}
	delay := RestartBackoffBase
	for i := 1; i < failures; i++ {
		if delay >= RestartBackoffMax/2 {
			return RestartBackoffMax
		}
		delay *= 2
	}
	if delay > RestartBackoffMax {
		return RestartBackoffMax
	}
	return delay
}

// BackoffTime returns the instant after which a restart is allowed.
func BackoffTime(failures int, now time.Time) time.Time {
	return now.Add(BackoffDelay(failures))
}
