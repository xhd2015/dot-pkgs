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
	if snap.Source != "billing" {
		t.Fatalf("source = %q", snap.Source)
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
