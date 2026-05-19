package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ProcessData maneja el cifrado o descifrado de bytes crudos
func ProcessData(datos []byte, key []byte, mode string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	switch mode {
	case "encrypt":
		nonce := make([]byte, aesGCM.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}
		return aesGCM.Seal(nonce, nonce, datos, nil), nil
		case "decrypt":
		nonceSize := aesGCM.NonceSize()
		if len(datos) < nonceSize {
			return nil, fmt.Errorf("data is too short or corrupted")
		}
		nonce, ciphertext := datos[:nonceSize], datos[nonceSize:]
		resultado, err := aesGCM.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, fmt.Errorf("decryption error (wrong password?): %v", err)
		}
		return resultado, nil
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}

// ProcessFile lee, procesa y guarda un archivo
func ProcessFile(inPath string, outPath string, key []byte, mode string) error {
	datos, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", inPath, err)
	}

	resultado, err := ProcessData(datos, key, mode)
	if err != nil {
		if strings.Contains(err.Error(), "too short") {
			return fmt.Errorf("the file %s is too short or corrupted", inPath)
		}
		if strings.Contains(err.Error(), "decryption error") {
			return fmt.Errorf("error decrypting %s (wrong password?): %v", inPath, err)
		}
		return err
	}

	// Asegurar que el directorio de destino exista antes de escribir
	err = os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory: %v", err)
	}

	err = os.WriteFile(outPath, resultado, 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", outPath, err)
	}

	return nil
}

// AdjustDestinationName agrega .enc al cifrar o quita .enc al descifrar
func AdjustDestinationName(baseDest string, mode string) string {
	if mode == "encrypt" && !strings.HasSuffix(baseDest, ".enc") {
		return baseDest + ".enc"
	}
	if mode == "decrypt" && strings.HasSuffix(baseDest, ".enc") {
		return strings.TrimSuffix(baseDest, ".enc")
	}
	return baseDest
}
