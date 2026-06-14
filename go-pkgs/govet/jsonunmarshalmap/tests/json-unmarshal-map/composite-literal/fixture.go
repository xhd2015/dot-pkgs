package main
import "encoding/json"
func main() {
	var data []byte
	json.Unmarshal(data, &map[string]any{})
}
