package p

var globalObj *int

func process(obj *int) int { return *obj }

func f() int {
	obj := globalObj
	if obj != nil {
		return process(obj)
	}
	return 0
}
