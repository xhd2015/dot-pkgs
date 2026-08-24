package applescript

import "testing"

func TestWriteTextExpr_Empty(t *testing.T) {
	if got := WriteTextExpr(""); got != `""` {
		t.Fatalf("got %q", got)
	}
}

func TestWriteTextExpr_Printable(t *testing.T) {
	if got := WriteTextExpr(`say "hi"`); got != `"say \"hi\""` {
		t.Fatalf("got %q", got)
	}
}

func TestWriteTextExpr_CtrlC(t *testing.T) {
	if got := WriteTextExpr("\x03"); got != `(ASCII character 3)` {
		t.Fatalf("got %q", got)
	}
}

func TestWriteTextExpr_CSIUp(t *testing.T) {
	want := `(ASCII character 27) & "[A"`
	if got := WriteTextExpr("\x1b[A"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWriteTextExpr_Mixed(t *testing.T) {
	want := `"hi" & (ASCII character 3) & "!"`
	if got := WriteTextExpr("hi\x03!"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
