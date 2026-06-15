package p

import "strconv"

func f() int {
	_, myErr := strconv.Atoi("42")
	if myErr != nil {
		return 0
	}
	return 0
}
