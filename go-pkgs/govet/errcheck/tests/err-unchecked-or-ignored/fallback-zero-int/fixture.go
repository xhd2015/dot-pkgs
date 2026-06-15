package p

import "strconv"

func f() int {
	n, err := strconv.Atoi("42")
	if err != nil {
		return 0
	}
	return n
}
