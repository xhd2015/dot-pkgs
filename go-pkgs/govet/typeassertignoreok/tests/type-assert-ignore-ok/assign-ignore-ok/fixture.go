package main
func main() {
	var x interface{} = "hello"
	var val string
	val, _ = x.(string)
	_ = val
}
