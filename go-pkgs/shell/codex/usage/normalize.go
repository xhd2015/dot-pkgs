package usage

import (
	"encoding/json"
	"fmt"
	"time"
)

// Snapshot is the normalized Codex account usage view for callers.
type Snapshot struct {
	PlanType         string
	RemainingPercent int // -1 when unknown
	UsedPercent      int // -1 when unknown
	ResetAt          time.Time
	Source           string // "codex/usage" | "wham/usage"
	Email            string // may be empty; fixtures use user@example.com
}

type usagePayload struct {
	PlanType string `json:"plan_type"`
	Email    string `json:"email"`
	Credits  *struct {
		HasCredits          bool `json:"has_credits"`
		Unlimited           bool `json:"unlimited"`
		OverageLimitReached bool `json:"overage_limit_reached"`
	} `json:"credits"`
	RateLimit *struct {
		PrimaryWindow *struct {
			UsedPercent        int   `json:"used_percent"`
			LimitWindowSeconds int   `json:"limit_window_seconds"`
			ResetAfterSeconds  int   `json:"reset_after_seconds"`
			ResetAt            int64 `json:"reset_at"`
		} `json:"primary_window"`
		SecondaryWindow *struct {
			UsedPercent        int   `json:"used_percent"`
			LimitWindowSeconds int   `json:"limit_window_seconds"`
			ResetAfterSeconds  int   `json:"reset_after_seconds"`
			ResetAt            int64 `json:"reset_at"`
		} `json:"secondary_window"`
	} `json:"rate_limit"`
	SpendControl *struct {
		IndividualLimit *struct {
			Limit             string `json:"limit"`
			Used              string `json:"used"`
			UsedPercent       int    `json:"used_percent"`
			RemainingPercent  int    `json:"remaining_percent"`
			ResetAfterSeconds int64  `json:"reset_after_seconds"`
			ResetAt           int64  `json:"reset_at"`
		} `json:"individual_limit"`
	} `json:"spend_control"`
}

// NormalizeJSON maps a ChatGPT usage JSON body into Snapshot.
func NormalizeJSON(raw []byte, source string) (Snapshot, error) {
	if len(raw) == 0 {
		return Snapshot{}, fmt.Errorf("codex usage: empty json")
	}
	var p usagePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return Snapshot{}, fmt.Errorf("codex usage: decode json: %w", err)
	}
	out := Snapshot{
		PlanType:         p.PlanType,
		Email:            p.Email,
		RemainingPercent: -1,
		UsedPercent:      -1,
		Source:           source,
	}

	if p.SpendControl != nil && p.SpendControl.IndividualLimit != nil {
		lim := p.SpendControl.IndividualLimit
		out.UsedPercent = lim.UsedPercent
		out.RemainingPercent = lim.RemainingPercent
		if lim.ResetAt > 0 {
			out.ResetAt = time.Unix(lim.ResetAt, 0)
		}
		return out, nil
	}

	if p.RateLimit != nil && p.RateLimit.PrimaryWindow != nil {
		w := p.RateLimit.PrimaryWindow
		out.UsedPercent = w.UsedPercent
		out.RemainingPercent = 100 - w.UsedPercent
		if out.RemainingPercent < 0 {
			out.RemainingPercent = 0
		}
		if w.ResetAt > 0 {
			out.ResetAt = time.Unix(w.ResetAt, 0)
		}
		return out, nil
	}

	return Snapshot{}, fmt.Errorf("codex usage: no spend_control or rate_limit windows in response")
}
