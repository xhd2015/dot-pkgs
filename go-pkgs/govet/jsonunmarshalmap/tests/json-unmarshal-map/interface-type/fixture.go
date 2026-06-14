package main
import "encoding/json"
func main() {
	var data []byte
	var m map[string]interface{}
	json.Unmarshal(data, &m)
}
