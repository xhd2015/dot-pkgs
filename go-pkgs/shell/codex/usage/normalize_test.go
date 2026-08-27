package usage

import (
	"testing"
	"time"
)

func TestNormalizeJSON_SpendControl(t *testing.T) {
	raw := []byte(`{
  "plan_type": "business",
  "email": "user@example.com",
  "rate_limit": null,
  "spend_control": {
    "individual_limit": {
      "limit": "15000",
      "used": "9338.8",
      "used_percent": 62,
      "remaining_percent": 38,
      "reset_after_seconds": 394827,
      "reset_at": 1788220800
    }
  }
}`)
	snap, err := NormalizeJSON(raw, "codex/usage")
	if err != nil {
		t.Fatal(err)
	}
	if snap.PlanType != "business" || snap.RemainingPercent != 38 || snap.UsedPercent != 62 {
		t.Fatalf("snap = %+v", snap)
	}
	if snap.Email != "user@example.com" || snap.Source != "codex/usage" {
		t.Fatalf("snap = %+v", snap)
	}
	want := time.Unix(1788220800, 0)
	if !snap.ResetAt.Equal(want) {
		t.Fatalf("reset = %v want %v", snap.ResetAt, want)
	}
}

func TestNormalizeJSON_PrimaryWindow(t *testing.T) {
	raw := []byte(`{
  "plan_type": "plus",
  "email": "user@example.com",
  "rate_limit": {
    "primary_window": {
      "used_percent": 25,
      "limit_window_seconds": 18000,
      "reset_at": 1730947200
    }
  }
}`)
	snap, err := NormalizeJSON(raw, "wham/usage")
	if err != nil {
		t.Fatal(err)
	}
	if snap.UsedPercent != 25 || snap.RemainingPercent != 75 || snap.Source != "wham/usage" {
		t.Fatalf("snap = %+v", snap)
	}
}

func TestNormalizeJSON_EmptyOrMissing(t *testing.T) {
	if _, err := NormalizeJSON(nil, "codex/usage"); err == nil {
		t.Fatal("want empty error")
	}
	if _, err := NormalizeJSON([]byte(`{"plan_type":"business"}`), "codex/usage"); err == nil {
		t.Fatal("want missing windows error")
	}
}
