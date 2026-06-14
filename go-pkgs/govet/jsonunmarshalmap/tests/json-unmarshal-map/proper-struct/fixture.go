package main
import "encoding/json"
type MyStruct struct {
	Name string
}
func main() {
	var data []byte
	var s MyStruct
	json.Unmarshal(data, &s)
}
