package main
import "encoding/json"
func main() {
	var data []byte
	var m1 map[string]any
	var m2 map[string]any
	json.Unmarshal(data, &m1)
	json.Unmarshal(data, &m2)
}
