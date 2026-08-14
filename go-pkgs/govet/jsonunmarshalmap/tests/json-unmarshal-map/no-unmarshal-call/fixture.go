package main
import "encoding/json"
func main() {
	var m map[string]any
	_ = m
	_ = json.RawMessage(nil)
}
