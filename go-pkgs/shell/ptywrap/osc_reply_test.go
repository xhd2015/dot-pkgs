package ptywrap

import (
	"bytes"
	"strings"
	"testing"
)

func TestConsumeOSCColorQueries_BEL(t *testing.T) {
	q := []byte("\x1b]11;?\x07")
	replies, rest := consumeOSCColorQueries(q)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Contains(replies, []byte("\x1b]11;rgb:")) {
		t.Fatalf("reply missing OSC 11 rgb: %q", replies)
	}
	if !bytes.HasSuffix(replies, []byte{0x07}) {
		t.Fatalf("reply should end BEL: %q", replies)
	}
}

func TestConsumeOSCColorQueries_ST(t *testing.T) {
	q := []byte("\x1b]10;?\x1b\\")
	replies, rest := consumeOSCColorQueries(q)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !bytes.Contains(replies, []byte("\x1b]10;rgb:")) {
		t.Fatalf("reply missing OSC 10 rgb: %q", replies)
	}
}

func TestConsumeOSCColorQueries_SplitAcrossChunks(t *testing.T) {
	var partial []byte
	var got []byte
	write := func(b []byte) error {
		got = append(got, b...)
		return nil
	}
	partial = maybeAutoReplyOSC(write, partial, []byte("\x1b]11;"))
	if len(got) != 0 {
		t.Fatalf("premature reply: %q", got)
	}
	partial = maybeAutoReplyOSC(write, partial, []byte("?\x07"))
	if !bytes.Contains(got, []byte("\x1b]11;rgb:")) {
		t.Fatalf("expected reply after second chunk, got %q partial=%q", got, partial)
	}
}

func TestConsumeOSCColorQueries_IgnoresOSC0(t *testing.T) {
	q := []byte("\x1b]0;title\x07hello")
	replies, rest := consumeOSCColorQueries(q)
	if len(replies) != 0 {
		t.Fatalf("should not reply to OSC 0: %q", replies)
	}
	if len(rest) != 0 {
		t.Fatalf("rest should be empty after complete non-query: %q", rest)
	}
}

func TestMaybeAutoReplyOSC_Disabled(t *testing.T) {
	t.Setenv("PTYWRAP_NO_OSC_REPLY", "1")
	var got []byte
	rest := maybeAutoReplyOSC(func(b []byte) error {
		got = append(got, b...)
		return nil
	}, nil, []byte("\x1b]11;?\x07"))
	if len(got) != 0 || rest != nil {
		t.Fatalf("disabled should no-op, got=%q rest=%q", got, rest)
	}
}

func TestConsumeOSCColorQueries_EmbeddedInNoise(t *testing.T) {
	q := []byte("noise\x1b]11;?\x07more")
	replies, rest := consumeOSCColorQueries(q)
	if len(rest) != 0 {
		t.Fatalf("rest=%q", rest)
	}
	if !strings.Contains(string(replies), "rgb:") {
		t.Fatalf("expected reply: %q", replies)
	}
}
