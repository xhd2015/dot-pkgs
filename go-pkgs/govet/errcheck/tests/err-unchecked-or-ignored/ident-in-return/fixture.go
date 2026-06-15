package p

var globalObj *int

func f() int {
	obj := globalObj
	if obj != nil {
		return *obj
	}
	return 0
}
