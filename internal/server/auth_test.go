package server

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor(t *testing.T) {
	authorizedKeys := []string{"key-1", "key-2"}
	interceptor := NewAuthInterceptor(authorizedKeys)
	unary := interceptor.Unary()

	// Handler dummy
	dummyHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	tests := []struct {
		name       string
		method     string
		md         metadata.MD
		wantStatus codes.Code
	}{
		{
			name:       "Access granted with correct x-api-key",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         metadata.Pairs("x-api-key", "key-1"),
			wantStatus: codes.OK,
		},
		{
			name:       "Access granted with correct raw authorization header",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         metadata.Pairs("authorization", "key-2"),
			wantStatus: codes.OK,
		},
		{
			name:       "Access granted with correct Bearer token in authorization header",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         metadata.Pairs("authorization", "Bearer key-1"),
			wantStatus: codes.OK,
		},
		{
			name:       "Authentication bypassed for Reflection service",
			method:     "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
			md:         metadata.MD{},
			wantStatus: codes.OK,
		},
		{
			name:       "Authentication fails due to missing metadata",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         nil,
			wantStatus: codes.Unauthenticated,
		},
		{
			name:       "Authentication fails due to invalid api key",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         metadata.Pairs("x-api-key", "invalid-key"),
			wantStatus: codes.PermissionDenied,
		},
		{
			name:       "Authentication fails due to empty api key value",
			method:     "/cifratin.v1.CifratinService/EncryptFile",
			md:         metadata.Pairs("x-api-key", ""),
			wantStatus: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.md != nil {
				ctx = metadata.NewIncomingContext(context.Background(), tt.md)
			} else {
				ctx = context.Background()
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: tt.method,
			}

			_, err := unary(ctx, nil, info, dummyHandler)
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected status error, got: %v", err)
				}
				if st.Code() != tt.wantStatus {
					t.Errorf("Unary() error code = %v, want %v", st.Code(), tt.wantStatus)
				}
			} else if tt.wantStatus != codes.OK {
				t.Errorf("Unary() returned no error, want error code = %v", tt.wantStatus)
			}
		})
	}
}
