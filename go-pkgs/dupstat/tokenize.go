package dupstat

import (
	"go/scanner"
	"go/token"
)

type TokenStrategy string

const (
	StrategyRaw        TokenStrategy = "raw"
	StrategyNormalized TokenStrategy = "normalized"
	StrategyMixed      TokenStrategy = "mixed"
)

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

var goOperators = map[token.Token]string{
	token.ADD: "+", token.SUB: "-", token.MUL: "*", token.QUO: "/", token.REM: "%",
	token.AND: "&", token.OR: "|", token.XOR: "^", token.SHL: "<<", token.SHR: ">>",
	token.AND_NOT: "&^", token.LAND: "&&", token.LOR: "||", token.ARROW: "<-", token.INC: "++", token.DEC: "--",
	token.EQL: "==", token.LSS: "<", token.GTR: ">", token.ASSIGN: "=", token.NOT: "!",
	token.NEQ: "!=", token.LEQ: "<=", token.GEQ: ">=", token.DEFINE: ":=", token.ELLIPSIS: "...",
	token.LPAREN: "(", token.RPAREN: ")", token.LBRACK: "[", token.RBRACK: "]", token.LBRACE: "{", token.RBRACE: "}",
	token.COMMA: ",", token.PERIOD: ".", token.SEMICOLON: ";", token.COLON: ":",
}

func opStr(tok token.Token) string {
	if s, ok := goOperators[tok]; ok {
		return s
	}
	return tok.String()
}

func tokenize(src []byte) []string {
	var s scanner.Scanner
	fset := token.NewFileSet()
	f := fset.AddFile("", -1, len(src))
	s.Init(f, src, nil, scanner.ScanComments)

	var tokens []string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.COMMENT {
			continue
		}
		if tok == token.IDENT {
			tokens = append(tokens, lit)
		} else if tok == token.INT || tok == token.FLOAT || tok == token.IMAG || tok == token.CHAR || tok == token.STRING {
			tokens = append(tokens, lit)
		} else if tok == token.SEMICOLON {
			continue
		} else {
			tokens = append(tokens, opStr(tok))
		}
	}
	return tokens
}

func tokenizeRaw(src []byte) []string {
	return tokenize(src)
}

func tokenizeNormalized(src []byte) []string {
	rawTokens := tokenize(src)
	return normalizeTokens(rawTokens)
}

func tokenizeMixed(sigSrc, bodySrc []byte) []string {
	sigTokens := tokenizeRaw(sigSrc)
	bodyTokens := tokenizeNormalized(bodySrc)
	return append(sigTokens, bodyTokens...)
}

func normalizeTokens(tokens []string) []string {
	idMap := make(map[string]string)
	nextID := 1
	result := make([]string, len(tokens))
	for i, tok := range tokens {
		if goKeywords[tok] || isOperator(tok) || isLiteral(tok) {
			result[i] = tok
			continue
		}
		if _, ok := idMap[tok]; !ok {
			idMap[tok] = "$" + itoa(nextID)
			nextID++
		}
		result[i] = idMap[tok]
	}
	return result
}

func isOperator(s string) bool {
	for _, op := range goOperators {
		if op == s {
			return true
		}
	}
	return false
}

func isLiteral(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] == '"' || s[0] == '`' || s[0] == '\'' {
		return true
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
