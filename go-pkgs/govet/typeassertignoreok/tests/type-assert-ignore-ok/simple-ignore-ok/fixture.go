package main
func main() {
	var x interface{} = "hello"
	val, _ := x.(string)
	_ = val
}
