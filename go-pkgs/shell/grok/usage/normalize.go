package usage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Period type labels for Snapshot.PeriodType.
const (
	PeriodMonthly = "monthly"
	PeriodWeekly  = "weekly"
)

// Snapshot is the normalized Grok account billing/usage view for callers.
type Snapshot struct {
	Used             int64 // config.used.val (monthly payload); 0 when absent
	MonthlyLimit     int64 // config.monthlyLimit.val; 0 means no numeric monthly cap
	RemainingPercent int   // -1 when unknown
	UsedPercent      int   // -1 when unknown
	PeriodStart      time.Time
	PeriodEnd        time.Time
	ResetAt          time.Time // PeriodEnd when set
	PeriodType       string    // PeriodMonthly | PeriodWeekly | ""
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

		// Credits format (?format=credits).
		CreditUsagePercent *json.Number `json:"creditUsagePercent"`
		CurrentPeriod      *struct {
			Start string `json:"start"`
			End   string `json:"end"`
			Type  string `json:"type"`
		} `json:"currentPeriod"`
		ProductUsage []struct {
			Product      string      `json:"product"`
			UsagePercent json.Number `json:"usagePercent"`
		} `json:"productUsage"`
	} `json:"config"`
}

// NormalizeJSON maps a cli-chat-proxy /v1/billing body (monthly or credits) into Snapshot.
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
		v, err := numberInt64(p.Config.Used.Val)
		if err != nil {
			return Snapshot{}, fmt.Errorf("grok usage: used.val: %w", err)
		}
		out.Used = v
	}
	if p.Config.MonthlyLimit != nil {
		v, err := numberInt64(p.Config.MonthlyLimit.Val)
		if err != nil {
			return Snapshot{}, fmt.Errorf("grok usage: monthlyLimit.val: %w", err)
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
		out.PeriodType = PeriodMonthly
		pct := clampPercent(int((out.Used * 100) / out.MonthlyLimit))
		out.UsedPercent = pct
		out.RemainingPercent = 100 - pct
		if out.RemainingPercent < 0 {
			out.RemainingPercent = 0
		}
	}

	applyCreditsFields(&out, p)
	return out, nil
}

func applyCreditsFields(out *Snapshot, p billingPayload) {
	if p.Config == nil {
		return
	}
	// Monthly-capped payloads keep used/limit math; credits fields belong to the
	// separate ?format=credits response.
	if out.MonthlyLimit > 0 {
		return
	}
	if p.Config.CurrentPeriod != nil {
		if t, ok := parseTime(p.Config.CurrentPeriod.Start); ok {
			out.PeriodStart = t
		}
		if t, ok := parseTime(p.Config.CurrentPeriod.End); ok {
			out.PeriodEnd = t
			out.ResetAt = t
		}
		if typ := periodTypeFromAPI(p.Config.CurrentPeriod.Type); typ != "" {
			out.PeriodType = typ
		}
	}

	pct, ok := creditsUsedPercent(p)
	if !ok {
		return
	}
	out.UsedPercent = clampPercent(pct)
	out.RemainingPercent = 100 - out.UsedPercent
	if out.RemainingPercent < 0 {
		out.RemainingPercent = 0
	}
	if out.PeriodType == "" {
		out.PeriodType = PeriodWeekly
	}
}

func creditsUsedPercent(p billingPayload) (int, bool) {
	if p.Config == nil {
		return 0, false
	}
	if p.Config.CreditUsagePercent != nil {
		v, err := numberInt64(*p.Config.CreditUsagePercent)
		if err == nil {
			return int(v), true
		}
	}
	for _, pu := range p.Config.ProductUsage {
		if !strings.EqualFold(strings.TrimSpace(pu.Product), "GrokBuild") {
			continue
		}
		v, err := numberInt64(pu.UsagePercent)
		if err != nil {
			return 0, false
		}
		return int(v), true
	}
	if len(p.Config.ProductUsage) == 1 {
		v, err := numberInt64(p.Config.ProductUsage[0].UsagePercent)
		if err != nil {
			return 0, false
		}
		return int(v), true
	}
	return 0, false
}

func periodTypeFromAPI(s string) string {
	u := strings.ToUpper(strings.TrimSpace(s))
	switch {
	case strings.Contains(u, "WEEKLY"):
		return PeriodWeekly
	case strings.Contains(u, "MONTHLY"):
		return PeriodMonthly
	default:
		return ""
	}
}

// SelectPreferred chooses monthly when it has a numeric cap; otherwise weekly
// credits when a percent is known; otherwise the monthly snapshot (possibly uncapped).
func SelectPreferred(monthly, weekly Snapshot, monthlyOK, weeklyOK bool) (Snapshot, bool) {
	if monthlyOK && monthly.MonthlyLimit > 0 && monthly.UsedPercent >= 0 {
		out := monthly
		if out.PeriodType == "" {
			out.PeriodType = PeriodMonthly
		}
		return out, true
	}
	if weeklyOK && weekly.UsedPercent >= 0 {
		out := weekly
		if out.PeriodType == "" {
			out.PeriodType = PeriodWeekly
		}
		return out, true
	}
	if monthlyOK {
		out := monthly
		if out.PeriodType == "" && out.MonthlyLimit > 0 {
			out.PeriodType = PeriodMonthly
		}
		return out, true
	}
	if weeklyOK {
		out := weekly
		if out.PeriodType == "" {
			out.PeriodType = PeriodWeekly
		}
		return out, true
	}
	return Snapshot{}, false
}

func numberInt64(n json.Number) (int64, error) {
	v, err := n.Int64()
	if err == nil {
		return v, nil
	}
	f, ferr := n.Float64()
	if ferr != nil {
		return 0, err
	}
	return int64(f), nil
}

func clampPercent(pct int) int {
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
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
