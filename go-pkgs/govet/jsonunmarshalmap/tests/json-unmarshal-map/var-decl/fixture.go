package main
import "encoding/json"
func main() {
	var data []byte
	var m map[string]any
	json.Unmarshal(data, &m)
}
