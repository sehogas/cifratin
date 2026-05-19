package server

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor maneja la validación de tokens de API Key para autorizar llamadas.
type AuthInterceptor struct {
	validKeys map[string]bool
}

// NewAuthInterceptor inicializa el interceptor con una lista de API Keys autorizadas.
func NewAuthInterceptor(authorizedKeys []string) *AuthInterceptor {
	keysMap := make(map[string]bool)
	for _, key := range authorizedKeys {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			keysMap[trimmed] = true
		}
	}
	return &AuthInterceptor{validKeys: keysMap}
}

// Unary intercepta y valida llamadas gRPC unitarias.
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Omitimos la validación para el servicio de Reflection de gRPC
		// Esto facilita las pruebas de desarrollo con herramientas como Postman o grpcurl.
		if strings.HasPrefix(info.FullMethod, "/grpc.reflection.v1alpha.ServerReflection/") ||
			strings.HasPrefix(info.FullMethod, "/grpc.reflection.v1.ServerReflection/") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing request metadata")
		}

		// Intentamos obtener la clave desde el header "x-api-key" o "authorization"
		var apiKey string
		if keys := md.Get("x-api-key"); len(keys) > 0 {
			apiKey = keys[0]
		} else if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			// Soporta formato estándar "Bearer <token>" o token crudo
			parts := strings.Split(authHeaders[0], " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				apiKey = parts[1]
			} else {
				apiKey = authHeaders[0]
			}
		}

		if apiKey == "" {
			return nil, status.Error(codes.Unauthenticated, "API key is missing from headers")
		}

		if !a.validKeys[apiKey] {
			return nil, status.Error(codes.PermissionDenied, "unauthorized or invalid API key")
		}

		return handler(ctx, req)
	}
}
