package ptywrap

import (
	"bytes"
	"testing"
)

func TestStripOutputProbes_TermenvPair(t *testing.T) {
	// termenv writes OSC 11 query (ST) then CSI 6n.
	data := []byte("notice\x1b]11;?\x1b\\\x1b[6nwarning")
	display, rest := stripOutputProbes(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(display, []byte("noticewarning")) {
		t.Fatalf("display=%q want noticewarning", display)
	}
}

func TestStripOutputProbes_BELQuery(t *testing.T) {
	data := []byte("\x1b]11;?\x07hello")
	display, rest := stripOutputProbes(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(display, []byte("hello")) {
		t.Fatalf("display=%q", display)
	}
}

func TestStripOutputProbes_KeepsOSC0(t *testing.T) {
	data := []byte("\x1b]0;title\x07hello")
	display, rest := stripOutputProbes(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(display, data) {
		t.Fatalf("display=%q want original", display)
	}
}

func TestStripOutputProbes_SplitOSC(t *testing.T) {
	display, rest := stripOutputProbes(nil, []byte("hi\x1b]11;"), true, true)
	if !bytes.Equal(display, []byte("hi")) {
		t.Fatalf("first display=%q", display)
	}
	if !bytes.Equal(rest, []byte("\x1b]11;")) {
		t.Fatalf("first rest=%q", rest)
	}
	display, rest = stripOutputProbes(rest, []byte("?\x07more"), true, true)
	if len(rest) != 0 {
		t.Fatalf("second rest=%q", rest)
	}
	if !bytes.Equal(display, []byte("more")) {
		t.Fatalf("second display=%q", display)
	}
}

func TestStripOutputProbes_KillSwitchPassesThrough(t *testing.T) {
	data := []byte("\x1b]11;?\x07\x1b[6n")
	display, rest := stripOutputProbes(nil, data, false, false)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(display, data) {
		t.Fatalf("display=%q want original when strip disabled", display)
	}
}

func TestFilterInputReports_UserLeftoverPair(t *testing.T) {
	// Local terminal answers: OSC 11 report (ST) + CPR, then a real key.
	data := []byte("\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[20;1Rls\r")
	keep, rest := filterInputReports(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(keep, []byte("ls\r")) {
		t.Fatalf("keep=%q want ls\\r", keep)
	}
}

func TestFilterInputReports_BELReport(t *testing.T) {
	data := []byte("\x1b]11;rgb:1e1e/1e1e/1e1e\x07x")
	keep, rest := filterInputReports(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(keep, []byte("x")) {
		t.Fatalf("keep=%q", keep)
	}
}

func TestFilterInputReports_PreservesArrowsAndCtrlC(t *testing.T) {
	data := []byte("\x1b[A\x1b[B\x1b[C\x1b[D\x03")
	keep, rest := filterInputReports(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(keep, data) {
		t.Fatalf("keep=%q want original arrows/ctrl-c", keep)
	}
}

func TestFilterInputReports_SplitReportThenKeys(t *testing.T) {
	keep, rest := filterInputReports(nil, []byte("\x1b]11;rgb:0000"), true, true)
	if len(keep) != 0 {
		t.Fatalf("premature keep=%q", keep)
	}
	if !bytes.Equal(rest, []byte("\x1b]11;rgb:0000")) {
		t.Fatalf("rest=%q", rest)
	}
	keep, rest = filterInputReports(rest, []byte("/0000/0000\x1b\\ab"), true, true)
	if len(rest) != 0 {
		t.Fatalf("second rest=%q", rest)
	}
	if !bytes.Equal(keep, []byte("ab")) {
		t.Fatalf("keep=%q", keep)
	}
}

func TestFilterInputReports_KillSwitchPassesThrough(t *testing.T) {
	data := []byte("\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[20;1R")
	keep, rest := filterInputReports(nil, data, false, false)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(keep, data) {
		t.Fatalf("keep=%q want original when drop disabled", keep)
	}
}

func TestFilterInputReports_DoesNotDropCUP(t *testing.T) {
	data := []byte("\x1b[1;1H")
	keep, rest := filterInputReports(nil, data, true, true)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Equal(keep, data) {
		t.Fatalf("keep=%q want CUP passed through", keep)
	}
}
