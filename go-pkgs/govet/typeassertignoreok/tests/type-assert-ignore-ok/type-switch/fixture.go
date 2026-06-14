package main
func main() {
	var x interface{} = "hello"
	switch x.(type) {
	case string:
		_ = "string"
	}
}
