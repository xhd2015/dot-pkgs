package scan_repo

import (
	"context"
	"sync"
	"time"
)

// WarmRefreshMode controls how budgeted warm polish (unit refresh + walk-log
// consume) runs after warm serve.
//
// Zero value is WarmRefreshSync (today's blocking behavior).
type WarmRefreshMode int

const (
	// WarmRefreshSync runs polish inside Scan/ScanSession before return.
	// New discoveries may merge into this invocation's Result / OnRepo.
	WarmRefreshSync WarmRefreshMode = iota
	// WarmRefreshAsync returns after warm serve; polish runs in a background
	// job that only updates durable cache (repos.json / walk log). This
	// invocation's Result / OnRepo stay frozen at the serve snapshot.
	// Callers must Join (or Stop+Join) before process exit for min-budget wait.
	WarmRefreshAsync
)

// Session is the result of ScanSession: discovery Result plus an optional
// background refresh handle when WarmRefreshMode is Async.
type Session struct {
	Result Result
	// Refresh is non-nil only when async polish was started. Nil for Sync,
	// cold-only, NoCache, or when polish is disabled (negative budget and
	// nothing else to do). Join on a nil handle is a no-op via Session.Join.
	Refresh *RefreshHandle
}

// Join waits for background polish using the min-budget join rule.
// No-op when Refresh is nil.
func (s Session) Join(ctx context.Context) error {
	if s.Refresh == nil {
		return nil
	}
	return s.Refresh.Join(ctx)
}

// Stop requests early stop of background polish (no min-budget wait).
// No-op when Refresh is nil.
func (s Session) Stop() {
	if s.Refresh == nil {
		return
	}
	s.Refresh.Stop()
}

// RefreshHandle owns one background warm-polish job (unit refresh + walk-log
// consume for all warm roots of a ScanSession).
//
// Lifetime rule (work remaining):
//
//	run while work remains AND (now < start+budget OR main still running)
//
// Join marks "main finished":
//   - before budget → wait until budget or work done (unless Stop / Join ctx cancel)
//   - after budget → soft-stop soon
//
// Stop aborts min-budget wait; already-written index is kept.
// If no work remains, Join returns immediately (budget is max effort, not sleep).
type RefreshHandle struct {
	start  time.Time
	budget time.Duration // resolved wall budget; 0 if unit refresh disabled

	workerCancel context.CancelFunc
	done         chan struct{}

	joinOnce sync.Once
	stopOnce sync.Once
	joinCh   chan struct{} // closed on Join
	stopCh   chan struct{} // closed on Stop

	errMu sync.Mutex
	err   error
}

// Stop requests the polish worker to stop without waiting for the min budget.
// Already-persisted index updates are retained. Idempotent.
func (h *RefreshHandle) Stop() {
	if h == nil {
		return
	}
	h.stopOnce.Do(func() {
		close(h.stopCh)
		if h.workerCancel != nil {
			h.workerCancel()
		}
	})
}

// Join waits for the polish worker under the min-budget rule.
// If ctx is cancelled, Stop is invoked (abort min-budget wait) and Join still
// waits for the worker to exit, then returns ctx.Err() when the worker had no
// other error.
// Idempotent concurrent Join: all waiters block on done.
func (h *RefreshHandle) Join(ctx context.Context) error {
	if h == nil {
		return nil
	}
	h.joinOnce.Do(func() {
		close(h.joinCh)
		// If budget already elapsed, cancel worker immediately.
		// If still within budget, a watcher cancels at budgetDeadline unless
		// Stop fires first or worker finishes early.
		go h.watchJoinBudget()
	})

	select {
	case <-h.done:
		return h.getErr()
	case <-ctx.Done():
		h.Stop()
		<-h.done
		if err := h.getErr(); err != nil {
			return err
		}
		return ctx.Err()
	}
}

func (h *RefreshHandle) watchJoinBudget() {
	// Stop already cancelled worker.
	select {
	case <-h.stopCh:
		return
	default:
	}

	remaining := time.Until(h.start.Add(h.budget))
	if remaining <= 0 {
		if h.workerCancel != nil {
			h.workerCancel()
		}
		return
	}

	t := time.NewTimer(remaining)
	defer t.Stop()
	select {
	case <-t.C:
		if h.workerCancel != nil {
			h.workerCancel()
		}
	case <-h.stopCh:
		// Stop already cancelled.
	case <-h.done:
		// Worker finished (no work left or completed early).
	}
}

func (h *RefreshHandle) setErr(err error) {
	if err == nil {
		return
	}
	h.errMu.Lock()
	if h.err == nil {
		h.err = err
	}
	h.errMu.Unlock()
}

func (h *RefreshHandle) getErr() error {
	h.errMu.Lock()
	defer h.errMu.Unlock()
	return h.err
}

// polishJob is one warm root's durable-only polish work for the async worker.
type polishJob struct {
	absRoot   string
	cacheRoot string
	opts      Options
	ignore    ignoreConfig
	served    []Repo
}

// startAsyncRefresh launches a single worker that polishes jobs durable-only.
// budget is the resolved WarmRefreshBudget (0 if unit refresh disabled; walk-log
// consume may still run). start is the wall clock for min-budget join.
func startAsyncRefresh(budget time.Duration, jobs []polishJob) *RefreshHandle {
	if len(jobs) == 0 {
		return nil
	}
	workerCtx, workerCancel := context.WithCancel(context.Background())
	h := &RefreshHandle{
		start:        time.Now(),
		budget:       budget,
		workerCancel: workerCancel,
		done:         make(chan struct{}),
		joinCh:       make(chan struct{}),
		stopCh:       make(chan struct{}),
	}
	// When unit refresh is disabled, budget is 0: Join after start cancels
	// immediately via remaining<=0, which is correct for "stop when main ends
	// if past budget" (budget already "elapsed"). Walk-log may still run until
	// Join/Stop if it has sync budget.
	if budget <= 0 {
		h.budget = 0
	}

	go func() {
		defer close(h.done)
		defer workerCancel()
		gate := &refreshGate{workerCtx: workerCtx, extendPastBudget: true}
		for _, job := range jobs {
			if workerCtx.Err() != nil {
				return
			}
			// Durable only: onRepo nil so this Scan's Result is not mutated.
			_, stats, err := warmBudgetRefresh(workerCtx, job.absRoot, job.opts, job.cacheRoot, job.ignore, job.served, nil, gate)
			if job.opts.Debug {
				debugf(job.opts, "refresh async root=%s budget=%s eligible=%d refreshed=%d duration=%s",
					job.absRoot, stats.budget, stats.eligible, stats.refreshed, stats.duration)
			}
			if err != nil && workerCtx.Err() == nil {
				h.setErr(err)
				return
			}
			if workerCtx.Err() != nil {
				// Soft stop: partial index already written by warmBudgetRefresh.
				return
			}
			// Walk-log consume also durable-only.
			_, err = consumeWalkLog(workerCtx, job.cacheRoot, job.opts, job.absRoot, job.served, nil)
			if err != nil && workerCtx.Err() == nil {
				h.setErr(err)
				return
			}
			if workerCtx.Err() != nil {
				return
			}
		}
	}()
	return h
}
