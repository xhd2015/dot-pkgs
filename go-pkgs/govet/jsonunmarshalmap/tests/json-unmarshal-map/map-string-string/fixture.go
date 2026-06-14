package main
import "encoding/json"
func main() {
	var data []byte
	var m map[string]string
	json.Unmarshal(data, &m)
}
