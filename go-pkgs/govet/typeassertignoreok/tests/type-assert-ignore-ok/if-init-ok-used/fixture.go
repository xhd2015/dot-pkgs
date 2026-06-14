package main
func main() {
	var x interface{} = "hello"
	if val, ok := x.(string); ok {
		_ = val
	}
}
