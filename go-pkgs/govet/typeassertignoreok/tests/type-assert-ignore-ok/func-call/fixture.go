package main
func someFunc() (int, bool) { return 1, true }
func main() {
	val, _ := someFunc()
	_ = val
}
