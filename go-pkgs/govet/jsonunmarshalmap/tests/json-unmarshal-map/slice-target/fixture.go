package main
import "encoding/json"
func main() {
	var data []byte
	var s []string
	json.Unmarshal(data, &s)
}
