package main
import "encoding/json"
func main() {
	var data []byte
	m := map[string]any{}
	json.Unmarshal(data, &m)
}
