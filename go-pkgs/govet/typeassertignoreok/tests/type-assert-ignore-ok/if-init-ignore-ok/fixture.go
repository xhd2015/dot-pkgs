package main
func main() {
	var x interface{} = "hello"
	if val, _ := x.(string); true {
		_ = val
	}
}
