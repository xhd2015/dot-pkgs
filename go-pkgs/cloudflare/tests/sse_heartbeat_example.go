package tests

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		flusher.Flush()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
				case <-ticker.C:
					// Send a heartbeat event every 15 seconds to keep the connection alive (if cloudflare in the middle, it can disconnect the connection after 2 minutes idle)
					fmt.Fprintf(w, ": heartbeat\n\n")
					flusher.Flush()
				case <-r.Context().Done():
					return
				}
		}
	})
	http.ListenAndServe(":8080", nil)
}