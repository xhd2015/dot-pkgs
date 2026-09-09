package main

import "testing"

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		`hello`:             `'hello'`,
		`a b; c`:            `'a b; c'`,
		`export X=1; cmd`:   `'export X=1; cmd'`,
		`it's`:              `'it'\''s'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Fatalf("shellQuote(%q)=%q want %q", in, got, want)
		}
	}
}
