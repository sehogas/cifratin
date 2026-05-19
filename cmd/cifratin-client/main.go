package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sehogas/cifratin/internal/crypto"
	v1 "github.com/sehogas/cifratin/pkg/api/v1"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// processRemoteFile envía el archivo al servidor gRPC para cifrarlo/descifrarlo y guarda el resultado
func processRemoteFile(ctx context.Context, client v1.CifratinServiceClient, inPath string, outPath string, password string, mode string, apiKey string) error {
	datos, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", inPath, err)
	}

	req := &v1.FileRequest{
		Data:     datos,
		Password: password,
	}

	// Inyectar API Key en los metadatos de salida
	ctxWithAuth := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-api-key", apiKey))

	var res *v1.FileResponse
	switch mode {
	case "encrypt":
		res, err = client.EncryptFile(ctxWithAuth, req)
	case "decrypt":
		res, err = client.DecryptFile(ctxWithAuth, req)
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}

	if err != nil {
		return fmt.Errorf("error in gRPC call for %s: %v", inPath, err)
	}

	// Asegurar que el directorio de destino exista antes de escribir
	err = os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory: %v", err)
	}

	err = os.WriteFile(outPath, res.ProcessedData, 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", outPath, err)
	}

	return nil
}

var Version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the application version")
	mode := flag.String("mode", "", "Action: 'encrypt' or 'decrypt'")
	inPath := flag.String("in", "", "Single file, search pattern (e.g. '*.pdf') or source directory")
	outPath := flag.String("out", "", "Destination file or destination directory")
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	apiKeyFlag := flag.String("apikey", "", "API key to authenticate with the gRPC server")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cifratin gRPC Client version: %s\n", Version)
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

	// API Key resolución: flag, luego variable de entorno, luego valor por defecto
	apiKey := *apiKeyFlag
	if apiKey == "" {
		apiKey = os.Getenv("CIFRATIN_API_KEY")
	}
	if apiKey == "" {
		apiKey = "dev-key-123" // Coincide con la del servidor por defecto
	}

	// 1. Leer la contraseña sin mostrarla en pantalla para la encriptación AES
	fmt.Print("Enter security password for the file: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		return
	}
	fmt.Println() // Salto de línea después de presionar Enter
	password := string(bytePassword)

	// 2. Conectarse al servidor gRPC
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to the gRPC server at %s: %v", *addr, err)
	}
	defer conn.Close()

	client := v1.NewCifratinServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 3. Determinar si el origen es un archivo único, patrón o directorio
	matches, err := filepath.Glob(*inPath)
	if err != nil {
		fmt.Printf("Error evaluating input pattern: %v\n", err)
		return
	}

	if len(matches) == 0 {
		fmt.Printf("No files found for: %s\n", *inPath)
		return
	}

	// Si hay una sola coincidencia, verificamos si es directorio o archivo individual
	if len(matches) == 1 {
		origenInfo, err := os.Stat(matches[0])
		if err != nil {
			fmt.Printf("Error reading source: %v\n", err)
			return
		}

		// A. DIRECTORY MODE
		if origenInfo.IsDir() {
			fmt.Printf("[gRPC Client] Processing directory recursively: %s -> %s\n", matches[0], *outPath)
			err = filepath.WalkDir(matches[0], func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil // Ignoramos directorios
				}

				relPath, _ := filepath.Rel(matches[0], path)
				destFinal := filepath.Join(*outPath, relPath)
				destFinal = crypto.AdjustDestinationName(destFinal, *mode)

				fmt.Printf("[%s via gRPC] %s...\n", *mode, path)
				return processRemoteFile(ctx, client, path, destFinal, password, *mode, apiKey)
			})

			if err != nil {
				fmt.Printf("Error processing directory: %v\n", err)
			} else {
				fmt.Println("Directory processing completed successfully.")
			}
			return
		}

		// B. SINGLE FILE MODE
		fmt.Printf("[gRPC Client] Processing single file: %s\n", matches[0])

		// Si el destino es una carpeta existente, metemos el archivo adentro
		destFinal := *outPath
		if outInfo, err := os.Stat(*outPath); err == nil && outInfo.IsDir() {
			destFinal = filepath.Join(*outPath, filepath.Base(matches[0]))
		}

		destFinal = crypto.AdjustDestinationName(destFinal, *mode)
		err = processRemoteFile(ctx, client, matches[0], destFinal, password, *mode, apiKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("File processed successfully via gRPC.")
		}
		return
	}

	// C. PATTERN MODE (Multiple files, e.g: *.pdf)
	fmt.Printf("[gRPC Client] Processing multiple files (%d found) into directory: %s\n", len(matches), *outPath)

	err = os.MkdirAll(*outPath, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		return
	}

	for _, match := range matches {
		destFinal := filepath.Join(*outPath, filepath.Base(match))
		destFinal = crypto.AdjustDestinationName(destFinal, *mode)

		fmt.Printf("[%s via gRPC] %s...\n", *mode, match)
		err = processRemoteFile(ctx, client, match, destFinal, password, *mode, apiKey)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", match, err)
		}
	}
	fmt.Println("Multiple processing completed.")
}
