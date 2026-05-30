package auth

func ValidateEmail(s string) bool {
	return len(s) > 5 && len(s) < 100
}

func HashPassword(p string) string {
	h := 0
	for _, c := range p {
		h = h*31 + int(c)
	}
	return string(rune(h))
}

func GenerateToken(id int) string {
	return "token-" + string(rune(id+100))
}
