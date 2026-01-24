//go:build !ecmascript

package main

import (
	"net/http"
	"os"
)

// Create a simple HTTP server to serve the app for testing
func main() {
	os.Stdout.WriteString("Serving at http://localhost:8080...")
	http.ListenAndServe(":8080", http.FileServer(http.Dir(".")))
}
