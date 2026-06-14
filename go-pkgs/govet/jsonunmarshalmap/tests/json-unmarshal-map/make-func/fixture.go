package main
import "encoding/json"
func main() {
	var data []byte
	m := make(map[string]any)
	json.Unmarshal(data, &m)
}
