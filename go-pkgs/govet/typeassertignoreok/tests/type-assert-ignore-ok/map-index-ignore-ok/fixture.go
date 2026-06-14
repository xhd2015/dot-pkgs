package main
func main() {
	m := map[string]any{"key": "hello"}
	val, _ := m["key"].(string)
	_ = val
}
