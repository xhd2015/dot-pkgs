package usage

import (
	"testing"
	"time"
)

const fixtureBillingLimited = `{
  "config": {
    "monthlyLimit": {"val": 100},
    "used": {"val": 73},
    "onDemandCap": {"val": 0},
    "billingPeriodStart": "2026-08-01T00:00:00+00:00",
    "billingPeriodEnd": "2026-09-01T00:00:00+00:00",
    "history": []
  }
}`

const fixtureBillingUnlimited = `{
  "config": {
    "monthlyLimit": {"val": 0},
    "used": {"val": 73},
    "billingPeriodStart": "2026-08-01T00:00:00+00:00",
    "billingPeriodEnd": "2026-09-01T00:00:00+00:00"
  }
}`

const fixtureBillingCreditsWeekly = `{
  "config": {
    "billingPeriodStart": "2026-08-28T00:55:25.179446+00:00",
    "billingPeriodEnd": "2026-09-04T00:55:25.179446+00:00",
    "creditUsagePercent": 2.0,
    "currentPeriod": {
      "start": "2026-08-28T00:55:25.179446+00:00",
      "end": "2026-09-04T00:55:25.179446+00:00",
      "type": "USAGE_PERIOD_TYPE_WEEKLY"
    },
    "productUsage": [
      {"product": "GrokBuild", "usagePercent": 2.0}
    ]
  }
}`

func TestNormalizeJSON_Limited(t *testing.T) {
	snap, err := NormalizeJSON([]byte(fixtureBillingLimited), "billing")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Used != 73 || snap.MonthlyLimit != 100 {
		t.Fatalf("snap = %+v", snap)
	}
	if snap.UsedPercent != 73 || snap.RemainingPercent != 27 {
		t.Fatalf("percents = %+v", snap)
	}
	if snap.Source != "billing" || snap.PeriodType != PeriodMonthly {
		t.Fatalf("source/period = %+v", snap)
	}
	wantEnd := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !snap.ResetAt.Equal(wantEnd) {
		t.Fatalf("ResetAt = %v", snap.ResetAt)
	}
}

func TestNormalizeJSON_Unlimited(t *testing.T) {
	snap, err := NormalizeJSON([]byte(fixtureBillingUnlimited), "billing")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Used != 73 || snap.MonthlyLimit != 0 {
		t.Fatalf("snap = %+v", snap)
	}
	if snap.UsedPercent != -1 || snap.RemainingPercent != -1 {
		t.Fatalf("want unknown percents, got %+v", snap)
	}
}

func TestNormalizeJSON_CreditsWeekly(t *testing.T) {
	snap, err := NormalizeJSON([]byte(fixtureBillingCreditsWeekly), "billing")
	if err != nil {
		t.Fatal(err)
	}
	if snap.UsedPercent != 2 || snap.RemainingPercent != 98 {
		t.Fatalf("percents = %+v", snap)
	}
	if snap.PeriodType != PeriodWeekly {
		t.Fatalf("PeriodType = %q", snap.PeriodType)
	}
	wantEnd := time.Date(2026, 9, 4, 0, 55, 25, 179446000, time.UTC)
	if !snap.ResetAt.Equal(wantEnd) {
		t.Fatalf("ResetAt = %v want %v", snap.ResetAt, wantEnd)
	}
}

func TestNormalizeJSON_Errors(t *testing.T) {
	if _, err := NormalizeJSON(nil, "billing"); err == nil {
		t.Fatal("want empty error")
	}
	if _, err := NormalizeJSON([]byte(`{}`), "billing"); err == nil {
		t.Fatal("want missing config")
	}
	if _, err := NormalizeJSON([]byte(`not-json`), "billing"); err == nil {
		t.Fatal("want decode error")
	}
}

func TestSelectPreferred(t *testing.T) {
	monthlyCapped := Snapshot{MonthlyLimit: 100, UsedPercent: 73, PeriodType: PeriodMonthly}
	monthlyOpen := Snapshot{MonthlyLimit: 0, Used: 73, UsedPercent: -1}
	weekly := Snapshot{UsedPercent: 2, PeriodType: PeriodWeekly}

	got, ok := SelectPreferred(monthlyCapped, weekly, true, true)
	if !ok || got.UsedPercent != 73 || got.PeriodType != PeriodMonthly {
		t.Fatalf("capped monthly should win: %+v ok=%v", got, ok)
	}

	got, ok = SelectPreferred(monthlyOpen, weekly, true, true)
	if !ok || got.UsedPercent != 2 || got.PeriodType != PeriodWeekly {
		t.Fatalf("weekly should win when monthly uncapped: %+v ok=%v", got, ok)
	}

	got, ok = SelectPreferred(monthlyOpen, Snapshot{}, true, false)
	if !ok || got.Used != 73 || got.UsedPercent != -1 {
		t.Fatalf("uncapped monthly fallback: %+v ok=%v", got, ok)
	}

	_, ok = SelectPreferred(Snapshot{}, Snapshot{}, false, false)
	if ok {
		t.Fatal("want no selection")
	}
}
