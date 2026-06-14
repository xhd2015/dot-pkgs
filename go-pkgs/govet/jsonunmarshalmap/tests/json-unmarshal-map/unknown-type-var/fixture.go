package main
import "encoding/json"
func someFunc() interface{} { return nil }
func main() {
	var data []byte
	m := someFunc()
	json.Unmarshal(data, &m)
}
