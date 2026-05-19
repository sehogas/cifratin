# ==============================================================================
# MAKEFILE DE CIFRATIN
# ==============================================================================
# Este Makefile agrupa todas las tareas comunes de desarrollo, pruebas,
# compilación y ejecución del proyecto Cifratin.
# Diseñado bajo principios de simplicidad (KISS) y automatización para 
# desarrolladores en entornos Windows y Linux.
# ==============================================================================

# Variables generales
BINARY_NAME=cifratin
BUILD_DIR=bin
CMD_PATH=./cmd/cifratin
CLIENT_BINARY_NAME=cifratin-client
CLIENT_CMD_PATH=./cmd/cifratin-client

.PHONY: help build clean test test-coverage run-cifrar-archivo run-descifrar-archivo run-cifrar-carpeta run-descifrar-carpeta run-cifrar-patron help-app

# Objetivo por defecto: muestra la ayuda general
help:
	@echo "======================================================================"
	@echo "                   OPCIONES DEL MAKEFILE DE CIFRATIN                  "
	@echo "======================================================================"
	@echo " Opciones de Desarrollo y Construccion:"
	@echo "   make build                   - Compila el binario CLI en la carpeta $(BUILD_DIR)"
	@echo "   make build-client            - Compila el binario Cliente gRPC en la carpeta $(BUILD_DIR)"
	@echo "   make build-server            - Compila el binario del servidor gRPC"
	@echo "   make clean                   - Elimina binarios y archivos temporales"
	@echo "   make test                    - Ejecuta todas las pruebas unitarias y de integracion"
	@echo "   make test-coverage           - Ejecuta las pruebas y muestra el reporte de cobertura"
	@echo ""
	@echo " Opciones de Ejecucion Directa (Cifrado / Descifrado):"
	@echo "   make run-cifrar-archivo      - Cifra un archivo de prueba de ejemplo (test.pdf)"
	@echo "   make run-descifrar-archivo   - Descifra el archivo de prueba cifrado"
	@echo "   make run-cifrar-carpeta      - Cifra el contenido de una carpeta de ejemplo (tests/)"
	@echo "   make run-descifrar-carpeta   - Descifra el contenido de la carpeta cifrada"
	@echo "   make run-cifrar-patron       - Cifra archivos usando un patron de busqueda (ej: tests/*.pdf)"
	@echo "   make help-app                - Muestra las instrucciones de uso nativas del CLI de Cifratin"
	@echo "======================================================================"

# ==============================================================================
# TAREAS DE COMPILACIÓN Y LIMPIEZA
# ==============================================================================

build:
	@echo "==> Compilando el binario CLI $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "==> Compilacion exitosa. Binario disponible en $(BUILD_DIR)/$(BINARY_NAME)"

build-server:
	@echo "==> Compilando el binario del servidor gRPC..."
	go build -o $(BUILD_DIR)/server ./cmd/server
	@echo "==> Compilacion exitosa. Binario disponible en $(BUILD_DIR)/server"

build-client:
	@echo "==> Compilando el binario Cliente gRPC $(CLIENT_BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) $(CLIENT_CMD_PATH)
	@echo "==> Compilacion exitosa. Binario disponible en $(BUILD_DIR)/$(CLIENT_BINARY_NAME)"

clean:
	@echo "==> Limpiando archivos generados y temporales..."
	go clean
	rm -rf $(BUILD_DIR)
	rm -rf test_output
	rm -f coverage.out
	@echo "==> Limpieza completada con exito."

# ==============================================================================
# TAREAS DE PRUEBAS (TESTING)
# ==============================================================================

test:
	@echo "==> Ejecutando pruebas unitarias y de integracion..."
	go test -v ./...

test-coverage:
	@echo "==> Calculando cobertura de codigo..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "==> (Opcional) Usa 'go tool cover -html=coverage.out' para ver el reporte detallado en el navegador."

# ==============================================================================
# TAREAS DE EJECUCIÓN PRÁCTICA (DEMOS Y CASOS DE USO)
# ==============================================================================
# IMPORTANTE: Al ejecutar cualquiera de estas opciones, la aplicacion solicitara
# la clave de seguridad de forma interactiva y segura en la consola mediante
# term.ReadPassword (sin mostrar los caracteres en pantalla).

run-cifrar-archivo:
	@echo "==> Ejecutando Cifratin en modo: Cifrar Archivo Individual..."
	@mkdir -p test_output
	go run $(CMD_PATH) -mode=encrypt -in=test.pdf -out=test_output/test.pdf.enc

run-descifrar-archivo:
	@echo "==> Ejecutando Cifratin en modo: Descifrar Archivo Individual..."
	@mkdir -p test_output
	go run $(CMD_PATH) -mode=decrypt -in=test_output/test.pdf.enc -out=test_output/test_descifrado.pdf

run-cifrar-carpeta:
	@echo "==> Ejecutando Cifratin en modo: Cifrar Carpeta (recursivo)..."
	@mkdir -p test_output/carpeta_enc
	go run $(CMD_PATH) -mode=encrypt -in=tests -out=test_output/carpeta_enc

run-descifrar-carpeta:
	@echo "==> Ejecutando Cifratin en modo: Descifrar Carpeta..."
	@mkdir -p test_output/carpeta_dec
	go run $(CMD_PATH) -mode=decrypt -in=test_output/carpeta_enc -out=test_output/carpeta_dec

run-cifrar-patron:
	@echo "==> Ejecutando Cifratin en modo: Cifrar por Patron (*.pdf)..."
	@mkdir -p test_output/patron_enc
	go run $(CMD_PATH) -mode=encrypt -in="tests/*.pdf" -out=test_output/patron_enc

help-app:
	@echo "==> Ayuda nativa del CLI de Cifratin:"
	go run $(CMD_PATH) -h

run-server:
	@echo "==> Levantando servidor gRPC..."
	go run ./cmd/server
