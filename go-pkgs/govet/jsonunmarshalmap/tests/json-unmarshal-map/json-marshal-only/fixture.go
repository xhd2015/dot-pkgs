package main
import "encoding/json"
func main() {
	m := map[string]any{"a": 1}
	json.Marshal(m)
}
