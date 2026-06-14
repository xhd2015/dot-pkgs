package main
import j "encoding/json"
func main() {
	var data []byte
	var m map[string]any
	j.Unmarshal(data, &m)
}
