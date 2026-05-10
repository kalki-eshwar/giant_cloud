package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/example/worker_node/daemon"
	"github.com/example/worker_node/jobs"
)

func main() {
	os.Setenv("GODEBUG", "http2server=0,http2client=0")
	// Initialize S3 backend
	jobs.InitS3()

	// Start the gRPC daemon client (replaces old background HTTP jobs)
	daemon.StartDaemon("localhost:8080", 100)

	// Keep an HTTP endpoint for receiving the actual large chunk payloads
	// Data plane remains HTTP/Streaming for large files, while control plane is gRPC
	http.HandleFunc("/chunk/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		// Store the uploaded encrypted chunk using the S3 storage job pipeline.
		job := &jobs.StorageJob{JobID: "job-123", ChunkID: "c1"}
		err = job.Execute(body)
		if err != nil {
			http.Error(w, "failed to store chunk", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "ok")
	})

	log.Println("Worker node starting on :8080")
	http.ListenAndServe(":8080", nil)
}
