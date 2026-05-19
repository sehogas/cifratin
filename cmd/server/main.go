package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sehogas/cifratin/internal/server"
	v1 "github.com/sehogas/cifratin/pkg/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var Version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the server version")
	port := flag.Int("port", 50051, "Port to listen for gRPC requests")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cifratin gRPC Server version: %s\n", Version)
		return
	}

	// Obtener claves de API autorizadas desde variables de entorno
	var keys []string
	envKeys := os.Getenv("CIFRATIN_API_KEYS")
	if envKeys != "" {
		keys = strings.Split(envKeys, ",")
		log.Printf("gRPC Server initialized with %d authorized API keys.", len(keys))
	} else {
		// Clave por defecto para desarrollo si no se especifica la variable de entorno
		defaultKey := "dev-key-123"
		keys = []string{defaultKey}
		log.Println("⚠️ WARNING: The CIFRATIN_API_KEYS environment variable is not configured.")
		log.Printf("A default API key has been generated for testing: %s", defaultKey)
	}

	authInterceptor := server.NewAuthInterceptor(keys)

	address := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to start tcp listener: %v", err)
	}

	// Registrar el interceptor de autenticación como UnaryInterceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	
	// Registramos nuestro servicio de cifrado
	cifratinServer := server.NewCryptoServiceServer()
	v1.RegisterCifratinServiceServer(grpcServer, cifratinServer)
	
	// Habilitamos reflection para facilitar pruebas con herramientas como grpcurl o Postman
	reflection.Register(grpcServer)

	log.Printf("Cifratin gRPC Server listening at %s...", address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failure: %v", err)
	}
}
