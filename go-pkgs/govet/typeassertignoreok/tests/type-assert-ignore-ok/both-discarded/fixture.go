package main
func main() {
	var x interface{} = "hello"
	_, _ = x.(string)
}
