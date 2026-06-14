package main
type Obj struct {
	Field interface{}
}
func main() {
	obj := Obj{Field: "hello"}
	val, _ := obj.Field.(string)
	_ = val
}
