package crypto

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

// ==============================================================================
// SUITE DE PRUEBAS UNITARIAS Y DE INTEGRACIÓN - CIFRATIN
// ==============================================================================
// Este archivo contiene las pruebas automatizadas para la lógica central de
// cifrado y descifrado (ProcesarArchivo) así como las funciones auxiliares
// (AjustarNombreDestino).
//
// Estrategia de Pruebas:
// - Pruebas Basadas en Tablas (Table-Driven Tests) para verificar múltiples
//   combinaciones de entrada/salida de forma limpia y mantenible (DRY).
// - Pruebas de Integración de Archivos utilizando t.TempDir() para garantizar
//   el aislamiento completo y evitar la persistencia de artefactos temporales.
// - Verificación de Casos de Éxito y Casos Borde (archivos corruptos, claves
//   incorrectas, rutas inexistentes y modos inválidos) para asegurar las
//   garantías de seguridad y robustez del sistema.
// ==============================================================================

/**
 * TestAjustarNombreDestino valida la correcta asignación de extensiones (.enc)
 * según el modo de operación (cifrar o descifrar) y la extensión actual del archivo.
 */
func TestAdjustDestinationName(t *testing.T) {
	casos := []struct {
		nombre   string
		baseDest string
		mode     string
		esperado string
	}{
		{
			nombre:   "Cifrar archivo sin extension enc",
			baseDest: "documento.pdf",
			mode:     "encrypt",
			esperado: "documento.pdf.enc",
		},
		{
			nombre:   "Cifrar archivo que ya tiene extension enc",
			baseDest: "backup.tar.enc",
			mode:     "encrypt",
			esperado: "backup.tar.enc",
		},
		{
			nombre:   "Descifrar archivo con extension enc",
			baseDest: "secreto.txt.enc",
			mode:     "decrypt",
			esperado: "secreto.txt",
		},
		{
			nombre:   "Descifrar archivo sin extension enc",
			baseDest: "archivosecreto.dat",
			mode:     "decrypt",
			esperado: "archivosecreto.dat",
		},
		{
			nombre:   "Modo desconocido no altera el nombre",
			baseDest: "data.json",
			mode:     "invalido",
			esperado: "data.json",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resultado := AdjustDestinationName(c.baseDest, c.mode)
			if resultado != c.esperado {
				t.Errorf("AdjustDestinationName(%q, %q) = %q; se esperaba %q", c.baseDest, c.mode, resultado, c.esperado)
			}
		})
	}
}

/**
 * TestProcesarArchivo evalúa el motor de cifrado y descifrado AES-256-GCM.
 * Cubre el ciclo completo de encriptación, desencriptación y validación de errores.
 */
func TestProcessFile(t *testing.T) {
	// 1. Configuración inicial y creación de entorno temporal aislado
	tempDir := t.TempDir()
	origenPath := filepath.Join(tempDir, "origen.txt")
	cifradoPath := filepath.Join(tempDir, "cifrado.enc")
	descifradoPath := filepath.Join(tempDir, "descifrado.txt")

	contenidoOriginal := []byte("Este es un mensaje confidencial de prueba para validar AES-256-GCM.")

	// Crear el archivo de origen
	err := os.WriteFile(origenPath, contenidoOriginal, 0644)
	if err != nil {
		t.Fatalf("Fallo al crear archivo de origen temporal: %v", err)
	}

	// Generar una clave aleatoria de 32 bytes (256 bits) para AES-256
	claveCorrecta := make([]byte, 32)
	_, err = rand.Read(claveCorrecta)
	if err != nil {
		t.Fatalf("Fallo al generar clave aleatoria: %v", err)
	}

	// Generar una clave incorrecta para pruebas de seguridad/autenticación
	claveIncorrecta := make([]byte, 32)
	_, err = rand.Read(claveIncorrecta)
	if err != nil {
		t.Fatalf("Fallo al generar clave incorrecta: %v", err)
	}

	// ==========================================================================
	// SUBTESTS DE COMPORTAMIENTO
	// ==========================================================================

	t.Run("1. Cifrar archivo exitosamente", func(t *testing.T) {
		err := ProcessFile(origenPath, cifradoPath, claveCorrecta, "encrypt")
		if err != nil {
			t.Fatalf("ProcessFile(encrypt) fallo inesperadamente: %v", err)
		}

		// Verificar que el archivo cifrado exista y contenga datos
		datosCifrados, err := os.ReadFile(cifradoPath)
		if err != nil {
			t.Fatalf("No se pudo leer el archivo cifrado generado: %v", err)
		}
		if len(datosCifrados) == 0 {
			t.Error("El archivo cifrado generado esta vacio")
		}
		if bytes.Equal(datosCifrados, contenidoOriginal) {
			t.Error("El archivo cifrado es identico al texto plano original (no se cifro)")
		}
	})

	t.Run("2. Descifrar archivo exitosamente", func(t *testing.T) {
		// Requiere que el subtest anterior haya generado cifradoPath exitosamente
		err := ProcessFile(cifradoPath, descifradoPath, claveCorrecta, "decrypt")
		if err != nil {
			t.Fatalf("ProcessFile(decrypt) fallo inesperadamente: %v", err)
		}

		// Verificar que el contenido descifrado coincida exactamente con el original
		datosDescifrados, err := os.ReadFile(descifradoPath)
		if err != nil {
			t.Fatalf("No se pudo leer el archivo descifrado: %v", err)
		}
		if !bytes.Equal(datosDescifrados, contenidoOriginal) {
			t.Errorf("Discordancia de datos.\nObtenido: %s\nEsperado: %s", string(datosDescifrados), string(contenidoOriginal))
		}
	})

	t.Run("3. Fallo al descifrar con clave incorrecta", func(t *testing.T) {
		descifradoInvalidoPath := filepath.Join(tempDir, "invalido.txt")
		err := ProcessFile(cifradoPath, descifradoInvalidoPath, claveIncorrecta, "decrypt")
		if err == nil {
			t.Error("Se esperaba un error al intentar descifrar con una clave incorrecta, pero no ocurrio")
		}
	})

	t.Run("4. Fallo al descifrar archivo corrupto o demasiado corto", func(t *testing.T) {
		cortoPath := filepath.Join(tempDir, "corto.enc")
		// Escribir un archivo con solo 5 bytes (menor que el tamaño del nonce de GCM, que es 12)
		err := os.WriteFile(cortoPath, []byte("12345"), 0644)
		if err != nil {
			t.Fatalf("Fallo al crear archivo corto: %v", err)
		}

		err = ProcessFile(cortoPath, filepath.Join(tempDir, "salida_corto.txt"), claveCorrecta, "decrypt")
		if err == nil {
			t.Error("Se esperaba un error al descifrar un archivo demasiado corto, pero no ocurrio")
		}
	})

	t.Run("5. Fallo por archivo de origen inexistente", func(t *testing.T) {
		noExistePath := filepath.Join(tempDir, "no_existe.txt")
		err := ProcessFile(noExistePath, filepath.Join(tempDir, "salida.enc"), claveCorrecta, "encrypt")
		if err == nil {
			t.Error("Se esperaba un error al procesar un archivo inexistente, pero no ocurrio")
		}
	})

	t.Run("6. Fallo por modo de operacion desconocido", func(t *testing.T) {
		err := ProcessFile(origenPath, filepath.Join(tempDir, "salida.dat"), claveCorrecta, "modoinvalido")
		if err == nil {
			t.Error("Se esperaba un error al utilizar un modo desconocido, pero no ocurrio")
		}
	})
}
