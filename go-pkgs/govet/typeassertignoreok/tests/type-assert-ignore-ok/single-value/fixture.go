package main
func main() {
	var x interface{} = "hello"
	val := x.(string)
	_ = val
}
