package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sehogas/cifratin/internal/crypto"
)

// FileRequest represents the JSON request payload for encryption and decryption.
type FileRequest struct {
	Data     []byte `json:"data"`
	Password string `json:"password"`
}

// FileResponse represents the JSON response payload.
type FileResponse struct {
	ProcessedData []byte `json:"processed_data"`
	Message       string `json:"message"`
}

// HTTPApiServer wraps the authorized API keys and handles request routing.
type HTTPApiServer struct {
	validKeys map[string]bool
}

// NewHTTPApiServer creates a new instance of HTTPApiServer.
func NewHTTPApiServer(authorizedKeys []string) *HTTPApiServer {
	keysMap := make(map[string]bool)
	for _, key := range authorizedKeys {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			keysMap[trimmed] = true
		}
	}
	return &HTTPApiServer{validKeys: keysMap}
}

// AuthMiddleware authenticates HTTP requests using API Keys.
// It checks for the "X-API-Key" header or "Authorization: Bearer <key>".
func (s *HTTPApiServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					apiKey = parts[1]
				} else {
					apiKey = authHeader
				}
			}
		}

		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "API key is missing from headers"}`))
			return
		}

		if !s.validKeys[apiKey] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "unauthorized or invalid API key"}`))
			return
		}

		next(w, r)
	}
}

// HandleEncrypt processes requests to encrypt data using AES-256-GCM.
func (s *HTTPApiServer) HandleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed. Use POST."}`))
		return
	}

	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error": "invalid JSON body: %v"}`, err)))
		return
	}

	if len(req.Data) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "no data provided"}`))
		return
	}
	if req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "password is required"}`))
		return
	}

	// Derive a 32-byte key from the password using SHA-256
	hash := sha256.Sum256([]byte(req.Password))
	key := hash[:]

	encryptedData, err := crypto.ProcessData(req.Data, key, "encrypt")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "encryption failed: %v"}`, err)))
		return
	}

	res := FileResponse{
		ProcessedData: encryptedData,
		Message:       "Data encrypted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// HandleDecrypt processes requests to decrypt data using AES-256-GCM.
func (s *HTTPApiServer) HandleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed. Use POST."}`))
		return
	}

	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error": "invalid JSON body: %v"}`, err)))
		return
	}

	if len(req.Data) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "no data provided"}`))
		return
	}
	if req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "password is required"}`))
		return
	}

	// Derive a 32-byte key from the password using SHA-256
	hash := sha256.Sum256([]byte(req.Password))
	key := hash[:]

	decryptedData, err := crypto.ProcessData(req.Data, key, "decrypt")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "decryption failed (wrong password?): %v"}`, err)))
		return
	}

	res := FileResponse{
		ProcessedData: decryptedData,
		Message:       "Data decrypted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
