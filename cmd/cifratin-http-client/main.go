package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sehogas/cifratin/internal/crypto"
	"golang.org/x/term"
)

// Version is the client version, injected dynamically during compilation.
var Version = "dev"

// FileRequest represents the JSON request payload sent to the HTTP server.
type FileRequest struct {
	Data     []byte `json:"data"`
	Password string `json:"password"`
}

// FileResponse represents the JSON response payload received from the HTTP server.
type FileResponse struct {
	ProcessedData []byte `json:"processed_data"`
	Message       string `json:"message"`
}

// processRemoteFileHTTP sends the file data to the HTTP server for encryption/decryption.
func processRemoteFileHTTP(client *http.Client, addr string, inPath string, outPath string, password string, mode string, apiKey string) error {
	datos, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", inPath, err)
	}

	reqPayload := FileRequest{
		Data:     datos,
		Password: password,
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("error marshalling JSON request: %v", err)
	}

	url := fmt.Sprintf("%s/%s", addr, mode)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var resPayload FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&resPayload); err != nil {
		return fmt.Errorf("failed to decode JSON response: %v", err)
	}

	// Ensure destination directory exists before writing
	err = os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory: %v", err)
	}

	err = os.WriteFile(outPath, resPayload.ProcessedData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to %s: %v", outPath, err)
	}

	return nil
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the application version")
	mode := flag.String("mode", "", "Action: 'encrypt' or 'decrypt'")
	inPath := flag.String("in", "", "Single file, search pattern (e.g. '*.pdf') or source directory")
	outPath := flag.String("out", "", "Destination file or destination directory")
	addr := flag.String("addr", "http://localhost:8080", "HTTP server address (including scheme)")
	apiKeyFlag := flag.String("apikey", "", "API key to authenticate with the HTTP server")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cifratin HTTP Client version: %s\n", Version)
		return
	}

	if *mode != "encrypt" && *mode != "decrypt" {
		fmt.Println("Error: You must specify -mode=encrypt or -mode=decrypt")
		return
	}
	if *inPath == "" || *outPath == "" {
		fmt.Println("Error: Missing parameters. Usage: app -mode=<encrypt|decrypt> -in=<source> -out=<destination>")
		return
	}

	// API Key resolution: flag, then env var, then default value
	apiKey := *apiKeyFlag
	if apiKey == "" {
		apiKey = os.Getenv("CIFRATIN_API_KEY")
	}
	if apiKey == "" {
		apiKey = "dev-key-123" // Matches the default server configuration
	}

	// Read the password interactively without echoing it on terminal
	fmt.Print("Enter security password for the file: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		return
	}
	fmt.Println() // Add a newline after Enter key
	password := string(bytePassword)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Determine if input is a single file, glob pattern, or folder
	matches, err := filepath.Glob(*inPath)
	if err != nil {
		fmt.Printf("Error evaluating input pattern: %v\n", err)
		return
	}

	if len(matches) == 0 {
		fmt.Printf("No files found for: %s\n", *inPath)
		return
	}

	if len(matches) == 1 {
		origenInfo, err := os.Stat(matches[0])
		if err != nil {
			fmt.Printf("Error reading source info: %v\n", err)
			return
		}

		// A. DIRECTORY MODE
		if origenInfo.IsDir() {
			fmt.Printf("[HTTP Client] Processing directory recursively: %s -> %s\n", matches[0], *outPath)
			err = filepath.WalkDir(matches[0], func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil // Skip directory nodes
				}

				relPath, _ := filepath.Rel(matches[0], path)
				destFinal := filepath.Join(*outPath, relPath)
				destFinal = crypto.AdjustDestinationName(destFinal, *mode)

				fmt.Printf("[%s via HTTP] %s...\n", *mode, path)
				return processRemoteFileHTTP(httpClient, *addr, path, destFinal, password, *mode, apiKey)
			})

			if err != nil {
				fmt.Printf("Error processing directory: %v\n", err)
			} else {
				fmt.Println("Directory processing completed successfully.")
			}
			return
		}

		// B. SINGLE FILE MODE
		fmt.Printf("[HTTP Client] Processing single file: %s\n", matches[0])

		destFinal := *outPath
		if outInfo, err := os.Stat(*outPath); err == nil && outInfo.IsDir() {
			destFinal = filepath.Join(*outPath, filepath.Base(matches[0]))
		}

		destFinal = crypto.AdjustDestinationName(destFinal, *mode)
		err = processRemoteFileHTTP(httpClient, *addr, matches[0], destFinal, password, *mode, apiKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("File processed successfully via HTTP.")
		}
		return
	}

	// C. PATTERN MODE (Globbing)
	fmt.Printf("[HTTP Client] Processing multiple files (%d found) into directory: %s\n", len(matches), *outPath)

	err = os.MkdirAll(*outPath, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		return
	}

	for _, match := range matches {
		destFinal := filepath.Join(*outPath, filepath.Base(match))
		destFinal = crypto.AdjustDestinationName(destFinal, *mode)

		fmt.Printf("[%s via HTTP] %s...\n", *mode, match)
		err = processRemoteFileHTTP(httpClient, *addr, match, destFinal, password, *mode, apiKey)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", match, err)
		}
	}
	fmt.Println("Multiple files processing completed.")
}
