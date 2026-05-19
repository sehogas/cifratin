package server

import (
	"context"
	"crypto/sha256"

	"github.com/sehogas/cifratin/internal/crypto"
	v1 "github.com/sehogas/cifratin/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CryptoServiceServer implements the gRPC CifratinServiceServer interface defined in the protobuf schema.
type CryptoServiceServer struct {
	v1.UnimplementedCifratinServiceServer
}

// NewCryptoServiceServer creates a new instance of CryptoServiceServer.
func NewCryptoServiceServer() *CryptoServiceServer {
	return &CryptoServiceServer{}
}

// EncryptFile encrypts the payload data received over gRPC using AES-256-GCM.
func (s *CryptoServiceServer) EncryptFile(ctx context.Context, req *v1.FileRequest) (*v1.FileResponse, error) {
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no data provided")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Hashear la contraseña para obtener 32 bytes (AES-256)
	hash := sha256.Sum256([]byte(req.Password))
	key := hash[:]

	encryptedData, err := crypto.ProcessData(req.Data, key, "encrypt")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encryption failed: %v", err)
	}

	return &v1.FileResponse{
		ProcessedData: encryptedData,
		Message:       "Data encrypted successfully",
	}, nil
}

// DecryptFile decrypts the payload data received over gRPC using AES-256-GCM.
func (s *CryptoServiceServer) DecryptFile(ctx context.Context, req *v1.FileRequest) (*v1.FileResponse, error) {
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no data provided")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Hashear la contraseña para obtener 32 bytes (AES-256)
	hash := sha256.Sum256([]byte(req.Password))
	key := hash[:]

	decryptedData, err := crypto.ProcessData(req.Data, key, "decrypt")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "decryption failed: %v", err)
	}

	return &v1.FileResponse{
		ProcessedData: decryptedData,
		Message:       "Data decrypted successfully",
	}, nil
}
