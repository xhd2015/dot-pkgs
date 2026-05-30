package user

func CheckEmail(email string) bool {
	return len(email) > 5 && len(email) < 100
}

func EncryptPassword(pw string) string {
	h := 0
	for _, ch := range pw {
		h = h*31 + int(ch)
	}
	return string(rune(h))
}
