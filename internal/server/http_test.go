package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPApiServerAuth(t *testing.T) {
	authorizedKeys := []string{"test-key-1", "test-key-2"}
	apiServer := NewHTTPApiServer(authorizedKeys)

	// Create dummy handler that simply returns status OK
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := apiServer.AuthMiddleware(dummyHandler)

	tests := []struct {
		name           string
		setupHeaders   func(req *http.Request)
		expectedStatus int
	}{
		{
			name: "Access granted with X-API-Key header",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("X-API-Key", "test-key-1")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Access granted with Authorization Bearer header",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-key-2")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Access granted with raw Authorization header",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Authorization", "test-key-1")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Access denied due to missing API Key",
			setupHeaders: func(req *http.Request) {
				// No headers set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Access denied due to invalid API Key",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("X-API-Key", "wrong-key")
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/encrypt", nil)
			tt.setupHeaders(req)
			rec := httptest.NewRecorder()

			protectedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHTTPApiServerEncryptDecrypt(t *testing.T) {
	authorizedKeys := []string{"key-123"}
	apiServer := NewHTTPApiServer(authorizedKeys)

	// Set up router to simulate calls
	mux := http.NewServeMux()
	mux.HandleFunc("/encrypt", apiServer.AuthMiddleware(apiServer.HandleEncrypt))
	mux.HandleFunc("/decrypt", apiServer.AuthMiddleware(apiServer.HandleDecrypt))

	server := httptest.NewServer(mux)
	defer server.Close()

	client := server.Client()

	plaintext := []byte("Sensitive data to be protected by Cifratin!")
	password := "supersecret"

	// 1. Test Encryption
	reqPayload := FileRequest{
		Data:     plaintext,
		Password: password,
	}
	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/encrypt", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "key-123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var encryptRes FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&encryptRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(encryptRes.ProcessedData) == 0 {
		t.Fatal("encrypted data is empty")
	}

	if bytes.Equal(encryptRes.ProcessedData, plaintext) {
		t.Fatal("encrypted data is identical to plaintext")
	}

	// 2. Test Decryption (Success)
	decryptPayload := FileRequest{
		Data:     encryptRes.ProcessedData,
		Password: password,
	}
	jsonBytesDec, err := json.Marshal(decryptPayload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	reqDec, err := http.NewRequest(http.MethodPost, server.URL+"/decrypt", bytes.NewBuffer(jsonBytesDec))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	reqDec.Header.Set("Content-Type", "application/json")
	reqDec.Header.Set("X-API-Key", "key-123")

	respDec, err := client.Do(reqDec)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer respDec.Body.Close()

	if respDec.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", respDec.StatusCode)
	}

	var decryptRes FileResponse
	if err := json.NewDecoder(respDec.Body).Decode(&decryptRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !bytes.Equal(decryptRes.ProcessedData, plaintext) {
		t.Errorf("decrypted data %q doesn't match original plaintext %q", string(decryptRes.ProcessedData), string(plaintext))
	}

	// 3. Test Decryption (Failure - Wrong Password)
	decryptPayloadWrong := FileRequest{
		Data:     encryptRes.ProcessedData,
		Password: "wrongpassword",
	}
	jsonBytesDecWrong, _ := json.Marshal(decryptPayloadWrong)

	reqDecWrong, _ := http.NewRequest(http.MethodPost, server.URL+"/decrypt", bytes.NewBuffer(jsonBytesDecWrong))
	reqDecWrong.Header.Set("Content-Type", "application/json")
	reqDecWrong.Header.Set("X-API-Key", "key-123")

	respDecWrong, err := client.Do(reqDecWrong)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer respDecWrong.Body.Close()

	if respDecWrong.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected decryption failure to return 500, got %d", respDecWrong.StatusCode)
	}
}

func TestHTTPApiServerValidation(t *testing.T) {
	authorizedKeys := []string{"key-123"}
	apiServer := NewHTTPApiServer(authorizedKeys)

	mux := http.NewServeMux()
	mux.HandleFunc("/encrypt", apiServer.AuthMiddleware(apiServer.HandleEncrypt))

	server := httptest.NewServer(mux)
	defer server.Close()

	client := server.Client()

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "Empty JSON body",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Password",
			payload: FileRequest{
				Data: []byte("some data"),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Data",
			payload: FileRequest{
				Password: "password",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, server.URL+"/encrypt", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "key-123")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
