package usage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Snapshot is the normalized Grok account billing/usage view for callers.
type Snapshot struct {
	Used             int64 // config.used.val
	MonthlyLimit     int64 // config.monthlyLimit.val; 0 means no numeric cap
	RemainingPercent int   // -1 when unknown (e.g. MonthlyLimit == 0)
	UsedPercent      int   // -1 when unknown
	PeriodStart      time.Time
	PeriodEnd        time.Time
	ResetAt          time.Time // PeriodEnd when set
	Source           string    // "billing"
	Email            string
}

type billingPayload struct {
	Config *struct {
		MonthlyLimit *struct {
			Val json.Number `json:"val"`
		} `json:"monthlyLimit"`
		Used *struct {
			Val json.Number `json:"val"`
		} `json:"used"`
		BillingPeriodStart string `json:"billingPeriodStart"`
		BillingPeriodEnd   string `json:"billingPeriodEnd"`
	} `json:"config"`
}

// NormalizeJSON maps a cli-chat-proxy /v1/billing body into Snapshot.
func NormalizeJSON(raw []byte, source string) (Snapshot, error) {
	if len(raw) == 0 {
		return Snapshot{}, fmt.Errorf("grok usage: empty json")
	}
	var p billingPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return Snapshot{}, fmt.Errorf("grok usage: decode json: %w", err)
	}
	if p.Config == nil {
		return Snapshot{}, fmt.Errorf("grok usage: missing config")
	}
	out := Snapshot{
		RemainingPercent: -1,
		UsedPercent:      -1,
		Source:           source,
	}
	if p.Config.Used != nil {
		v, err := p.Config.Used.Val.Int64()
		if err != nil {
			f, ferr := p.Config.Used.Val.Float64()
			if ferr != nil {
				return Snapshot{}, fmt.Errorf("grok usage: used.val: %w", err)
			}
			v = int64(f)
		}
		out.Used = v
	}
	if p.Config.MonthlyLimit != nil {
		v, err := p.Config.MonthlyLimit.Val.Int64()
		if err != nil {
			f, ferr := p.Config.MonthlyLimit.Val.Float64()
			if ferr != nil {
				return Snapshot{}, fmt.Errorf("grok usage: monthlyLimit.val: %w", err)
			}
			v = int64(f)
		}
		out.MonthlyLimit = v
	}
	if t, ok := parseTime(p.Config.BillingPeriodStart); ok {
		out.PeriodStart = t
	}
	if t, ok := parseTime(p.Config.BillingPeriodEnd); ok {
		out.PeriodEnd = t
		out.ResetAt = t
	}
	if out.MonthlyLimit > 0 {
		// Percent of limit used; clamp to [0, 100].
		pct := int((out.Used * 100) / out.MonthlyLimit)
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		out.UsedPercent = pct
		out.RemainingPercent = 100 - pct
		if out.RemainingPercent < 0 {
			out.RemainingPercent = 0
		}
	}
	return out, nil
}

func parseTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}
