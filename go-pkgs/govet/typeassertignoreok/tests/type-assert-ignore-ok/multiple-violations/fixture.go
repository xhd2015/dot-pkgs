package main
func main() {
	var a interface{} = "hello"
	var b interface{} = 42
	v1, _ := a.(string)
	v2, _ := b.(int)
	_, _ = v1, v2
}
