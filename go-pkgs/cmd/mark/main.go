package main

import "time"

// used on terminal, like:
//   $ mark "I'm still waiting for result" 
func main(){
	for {
	   time.Sleep(100*365*24*time.Hour)
	}
}