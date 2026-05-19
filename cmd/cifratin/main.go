package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sehogas/cifratin/internal/crypto"
	"golang.org/x/term"
)

var Version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the application version")
	mode := flag.String("mode", "", "Action: 'encrypt' or 'decrypt'")
	inPath := flag.String("in", "", "Single file, search pattern (e.g. '*.pdf') or source directory")
	outPath := flag.String("out", "", "Destination file or destination directory")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cifratin CLI version: %s\n", Version)
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

	// 1. Leer la contraseña sin mostrarla en pantalla
	fmt.Print("Enter security password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		return
	}
	fmt.Println() // Salto de línea después de que el usuario presiona Enter

	// Asegurar 32 bytes para AES-256 hasheando la contraseña
	hash := sha256.Sum256(bytePassword)
	key := hash[:]

	// 2. Determinar si el origen es un archivo único, patrón o directorio
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

		// A. MODO DIRECTORIO
		if origenInfo.IsDir() {
			fmt.Printf("Processing directory recursively: %s -> %s\n", matches[0], *outPath)
			err = filepath.WalkDir(matches[0], func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil // Ignoramos los directorios, solo procesamos sus archivos
				}

				relPath, _ := filepath.Rel(matches[0], path)
				destFinal := filepath.Join(*outPath, relPath)
				destFinal = crypto.AdjustDestinationName(destFinal, *mode)

				fmt.Printf("[%s] %s...\n", *mode, path)
				return crypto.ProcessFile(path, destFinal, key, *mode)
			})

			if err != nil {
				fmt.Printf("Error processing directory: %v\n", err)
			} else {
				fmt.Println("Directory processing completed successfully.")
			}
			return
		}

		// B. MODO ARCHIVO INDIVIDUAL
		fmt.Printf("Processing single file: %s\n", matches[0])

		// Si el destino pasado es una carpeta existente, lo metemos adentro
		destFinal := *outPath
		if outInfo, err := os.Stat(*outPath); err == nil && outInfo.IsDir() {
			destFinal = filepath.Join(*outPath, filepath.Base(matches[0]))
		}

		destFinal = crypto.AdjustDestinationName(destFinal, *mode)
		err = crypto.ProcessFile(matches[0], destFinal, key, *mode)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("File processed successfully.")
		}
		return
	}

	// C. MODO PATRÓN (Varios archivos, ej: *.pdf)
	fmt.Printf("Processing multiple files (%d found) into directory: %s\n", len(matches), *outPath)

	// Si hay varios archivos origen, el destino OBLIGATORIAMENTE debe tratarse como un directorio
	err = os.MkdirAll(*outPath, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		return
	}

	for _, match := range matches {
		destFinal := filepath.Join(*outPath, filepath.Base(match))
		destFinal = crypto.AdjustDestinationName(destFinal, *mode)

		fmt.Printf("[%s] %s...\n", *mode, match)
		err = crypto.ProcessFile(match, destFinal, key, *mode)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", match, err)
		}
	}
	fmt.Println("Multiple processing completed.")
}
