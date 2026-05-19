package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sehogas/cifratin/internal/server"
)

// Version is the server version, injected dynamically during compilation.
var Version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the server version")
	port := flag.Int("port", 8080, "Port to listen for HTTP requests")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cifratin HTTP Server version: %s\n", Version)
		return
	}

	// Retrieve authorized API keys from environment variables
	var keys []string
	envKeys := os.Getenv("CIFRATIN_API_KEYS")
	if envKeys != "" {
		keys = strings.Split(envKeys, ",")
		log.Printf("HTTP Server initialized with %d authorized API keys.", len(keys))
	} else {
		// Default key for local development
		defaultKey := "dev-key-123"
		keys = []string{defaultKey}
		log.Println("⚠️ WARNING: The CIFRATIN_API_KEYS environment variable is not configured.")
		log.Printf("A default API key has been generated for testing: %s", defaultKey)
	}

	apiServer := server.NewHTTPApiServer(keys)

	mux := http.NewServeMux()
	mux.HandleFunc("/encrypt", apiServer.AuthMiddleware(apiServer.HandleEncrypt))
	mux.HandleFunc("/decrypt", apiServer.AuthMiddleware(apiServer.HandleDecrypt))

	address := fmt.Sprintf(":%d", *port)
	log.Printf("Cifratin HTTP Server listening at http://localhost%s...", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("HTTP server failure: %v", err)
	}
}
