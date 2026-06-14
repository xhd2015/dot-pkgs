package main
func main() {
	var x interface{} = "hello"
	val, ok := x.(string)
	_, _ = val, ok
}
