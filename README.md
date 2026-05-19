# Cifratin 🔐

[![Go Version](https://img.shields.io/github/go-mod/go-version/sehogas/cifratin?color=00ADD8&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/sehogas/cifratin)](https://goreportcard.com/report/github.com/sehogas/cifratin)
[![Security Policy](https://img.shields.io/badge/Security-Policy_Enabled-success?logo=github-security)](SECURITY.md)
[![Latest Release](https://img.shields.io/github/v/release/sehogas/cifratin?logo=github)](https://github.com/sehogas/cifratin/releases)

Una utilidad de línea de comandos (CLI) escrita en Go diseñada para cifrar y descifrar archivos de forma rápida y segura. Implementa criptografía simétrica moderna utilizando el estándar **AES-256-GCM** y proporciona una interfaz limpia para procesar archivos individuales, patrones de búsqueda o directorios enteros de forma recursiva.

---

## 🛠️ Alcance y Características

El proyecto ha sido diseñado bajo los principios de simplicidad y solidez, asegurando que la gestión de datos confidenciales se realice de forma segura a nivel local.

- **Criptografía Robusta**: 
  - **Cifrado Simétrico**: Utiliza **AES-256** en modo **GCM** (Galois/Counter Mode). Este modo proporciona cifrado autenticado, lo que significa que garantiza tanto la **confidencialidad** como la **integridad** de los datos cifrados (detecta si el archivo ha sido alterado).
  - **Derivación de Clave**: Las contraseñas ingresadas por el usuario se procesan mediante **SHA-256** para derivar de forma consistente una clave simétrica de 32 bytes (256 bits), sin importar la longitud de la contraseña original.
- **Entrada Segura de Contraseñas**: La clave de seguridad se lee a través del terminal de forma interactiva usando `golang.org/x/term`, ocultando los caracteres ingresados (sin eco en pantalla) para evitar ataques de hombro (*shoulder surfing*).
- **Flexibilidad de Entrada**:
  - **Archivo Único**: Cifra o descifra un archivo individual especificando sus rutas.
  - **Patrones de Búsqueda (Wildcards)**: Permite utilizar expresiones como `*.pdf` o `tests/*.txt` para procesar múltiples archivos coincidentes en lote.
  - **Directorios (Procesamiento Recursivo)**: Recorre estructuras de carpetas completas y procesa cada archivo de forma individual, replicando la estructura de directorios en el destino correspondiente.
- **Servicio gRPC Integrado**: Incluye un servidor gRPC (`cmd/cifratin-server`) independiente del CLI para permitir la integración y ejecución del motor de cifrado AES desde clientes remotos o microservicios.
- **Manejo Inteligente de Extensiones**:
  - Agrega automáticamente la extensión `.enc` al cifrar archivos si no la poseen.
  - Remueve la extensión `.enc` al descifrar si se encuentra presente en el nombre del archivo origen.
- **Creación de Directorios**: Si el directorio destino de salida no existe en el sistema, la aplicación lo crea de forma automática antes de proceder con la escritura de los archivos.

---

## 🏗️ Requisitos Previos

Para compilar y ejecutar Cifratin en su máquina local, necesitará tener instalado:

- **Go**: Versión `1.25.1` o superior (según la definición del módulo en [go.mod](go.mod)).
- **Make**: Herramienta de automatización (opcional, pero altamente recomendada para ejecutar tareas del proyecto de forma simplificada).

> [!NOTE]
> **Modificaciones en la definición gRPC**: Si planea modificar los archivos `.proto` (dentro de `api/proto/v1/`), requerirá tener instalado en su sistema el compilador de Protocol Buffers (`protoc`) junto con los plugins de Go para la generación de código:
> - `protoc-gen-go` (instalable con `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`)
> - `protoc-gen-go-grpc` (instalable con `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`)
>
> Los archivos Go ya generados (`.pb.go`) se encuentran versionados en el repositorio, por lo que no es obligatorio instalar estas herramientas a menos que modifique la API del protocolo.

---

## 🚀 Instalación y Compilación

Para compilar el proyecto y generar el binario ejecutable, siga estos pasos:

# Compilar el binario ejecutable CLI
make build

# Compilar el cliente gRPC
make build-client

# Compilar el servidor gRPC
make build-server
```
El binario CLI resultante se guardará en `bin/cifratin`, el cliente en `bin/cifratin-client` y el servidor en `bin/cifratin-server` (o `.exe` en entornos Windows).

### Usando Go directamente
```bash
go build -o bin/cifratin ./cmd/cifratin
go build -o bin/cifratin-client ./cmd/cifratin-client
go build -o bin/cifratin-server ./cmd/cifratin-server
```

Si desea limpiar los archivos compilados y los resultados de las pruebas:
```bash
make clean
```

---

## 📖 Instrucciones de Uso

### 1. Cifrador Local (CLI Tradicional)

La aplicación CLI local recibe flags para parametrizar su comportamiento:

```bash
bin/cifratin -mode=<encrypt|decrypt> -in=<origen> -out=<destino>
```

---

### 2. Cifrador Remoto (Cliente gRPC)

La aplicación cliente gRPC replica la misma funcionalidad del CLI local, pero realiza las tareas criptográficas llamando remotamente al servidor gRPC.

```bash
bin/cifratin-client -mode=<encrypt|decrypt> -in=<origen> -out=<destino> [flags adicionales]
```

#### Flags Adicionales para el Cliente gRPC:
- `-addr`: Dirección IP y puerto del servidor gRPC. Por defecto es `localhost:50051`.
- `-apikey`: Clave de API para autenticarse con el servidor. Si no se especifica, intentará leer de la variable de entorno `CIFRATIN_API_KEY`, de lo contrario utilizará la clave por defecto `"dev-key-123"`.

> [!IMPORTANT]
> Tanto para la versión local como para la cliente gRPC, se le solicitará que ingrese la clave criptográfica para el archivo de forma interactiva en la consola.

---

### Ejemplos Prácticos de Ejecución

#### 1. Cifrado y Descifrado de un Archivo Individual (Local)

```bash
# Cifrar un archivo PDF individual
bin/cifratin -mode=encrypt -in=test.pdf -out=test.pdf.enc

# Descifrar el archivo cifrado anteriormente
bin/cifratin -mode=decrypt -in=test.pdf.enc -out=test_restaurado.pdf
```

#### 2. Cifrado y Descifrado de un Archivo Individual (Cliente gRPC)

Asegúrese de tener el servidor gRPC encendido (`make run-server`).

```bash
# Cifrar un archivo PDF individual enviándolo al servidor gRPC
bin/cifratin-client -mode=encrypt -in=test.pdf -out=test.pdf.enc

# Descifrar el archivo cifrado anteriormente a través de gRPC
bin/cifratin-client -mode=decrypt -in=test.pdf.enc -out=test_restaurado_grpc.pdf
```

#### 3. Cifrado y Descifrado de un Directorio Completo (Cliente gRPC)

El modo directorio en el cliente gRPC también mantiene la estructura recursiva interna de las carpetas.

```bash
# Cifrar todo el contenido de la carpeta 'tests' hacia 'salida_cifrada_grpc'
bin/cifratin-client -mode=encrypt -in=tests -out=salida_cifrada_grpc

# Descifrar el contenido cifrado hacia 'salida_restaurada_grpc'
bin/cifratin-client -mode=decrypt -in=salida_cifrada_grpc -out=salida_restaurada_grpc
```

---

## 🔒 Seguridad en el Servicio gRPC

Por defecto, la exposición de un servicio gRPC en una red puede ser vulnerable si no se restringe quién tiene permiso para invocarlo. Para mitigar esto, el servicio implementa una capa de autenticación basada en **API Keys** mediante un interceptor unitario (`UnaryServerInterceptor`).

### 1. Configuración del Servidor
Las claves válidas autorizadas se definen en el servidor a través de la variable de entorno `CIFRATIN_API_KEYS`, separadas por comas:

```bash
# En sistemas Windows (PowerShell)
$env:CIFRATIN_API_KEYS="mi-servicio-a,mi-servicio-b,api-gateway-key"
make run-server

# En sistemas Linux/macOS
export CIFRATIN_API_KEYS="mi-servicio-a,mi-servicio-b,api-gateway-key"
make run-server
```

> [!NOTE]
> Si arranca el servidor sin establecer esta variable, el sistema utilizará una clave por defecto (`dev-key-123`) para entornos de desarrollo y emitirá una advertencia en la consola.

### 2. Consumo desde los Clientes
Para que un servicio cliente pueda consumir con éxito los endpoints `EncryptFile` y `DecryptFile`, deberá adjuntar una clave válida en los metadatos de la llamada gRPC. El interceptor valida cualquiera de los siguientes dos encabezados:

- **Cabecera `x-api-key`**: Debe contener el token crudo (ej: `x-api-key: mi-servicio-a`).
- **Cabecera `authorization`**: Soporta tokens portadores estándar (ej: `authorization: Bearer mi-servicio-a`) o el valor crudo.

Cualquier llamada que carezca del encabezado o provea una clave inválida será rechazada a nivel de interceptor devolviendo los códigos de error gRPC `Unauthenticated` o `PermissionDenied` respectivamente.

---

### Ejemplos Prácticos de Ejecución

#### 1. Cifrado y Descifrado de un Archivo Individual

```bash
# Cifrar un archivo PDF individual
bin/cifratin -mode=encrypt -in=test.pdf -out=test.pdf.enc

# Descifrar el archivo cifrado anteriormente
bin/cifratin -mode=decrypt -in=test.pdf.enc -out=test_restaurado.pdf
```

#### 2. Cifrado y Descifrado de un Directorio Completo (Recursivo)

El modo directorio mantendrá la estructura interna de las carpetas al realizar la operación.

```bash
# Cifrar todo el contenido de la carpeta 'tests' hacia 'salida_cifrada'
bin/cifratin -mode=encrypt -in=tests -out=salida_cifrada

# Descifrar el contenido cifrado hacia 'salida_restaurada'
bin/cifratin -mode=decrypt -in=salida_cifrada -out=salida_restaurada
```

#### 3. Cifrado y Descifrado por Patrones (Globbing)

```bash
# Cifrar todos los archivos PDF dentro de la carpeta 'tests'
bin/cifratin -mode=encrypt -in="tests/*.pdf" -out=salida_patron
```

---

## 🛠️ Automatización del Desarrollo (Makefile)

El proyecto incluye un `Makefile` con comandos predefinidos para agilizar las tareas comunes de desarrollo y demostración:

| Comando | Descripción |
| :--- | :--- |
| `make help` | Muestra el menú de ayuda del Makefile con todas las opciones disponibles. |
| `make build` | Compila el código del CLI y genera el binario en la carpeta `bin/`. |
| `make build-server`| Compila el servidor gRPC y genera su ejecutable en `bin/`. |
| `make clean` | Elimina binarios compilados, coberturas y salidas temporales generadas. |
| `make test` | Corre las pruebas unitarias y de integración mostrando la salida en consola. |
| `make test-coverage` | Ejecuta las pruebas y calcula la cobertura del código. |
| `make run-cifrar-archivo` | Ejecuta una demostración de cifrado usando el archivo `test.pdf`. |
| `make run-descifrar-archivo`| Ejecuta una demostración de descifrado del archivo cifrado en la tarea anterior. |
| `make run-cifrar-carpeta` | Demostración práctica de cifrado recursivo de la carpeta `tests/`. |
| `make run-descifrar-carpeta`| Descifra la carpeta procesada recursivamente de regreso a texto plano. |
| `make run-cifrar-patron` | Demostración práctica de cifrado usando patrones de búsqueda de archivos. |
| `make run-server` | Inicia el servidor gRPC en el puerto 50051 (habilitado con reflection). |
| `make help-app` | Muestra las instrucciones de uso nativas del CLI impresas por Go. |

---

## 🧪 Pruebas y Cobertura de Código

El proyecto contiene un conjunto completo de pruebas unitarias y de integración que cubren los flujos de éxito y error principales, tales como:
- Ajustes automáticos de nombres de archivo de destino.
- Cifrado exitoso de datos individuales.
- Descifrado correcto utilizando la contraseña de origen.
- Manejo de fallos por claves incorrectas (verificación de integridad de AES-GCM).
- Manejo de archivos corruptos o incompletos.
- Comportamientos ante parámetros inválidos o rutas inexistentes.

### Ejecución de Pruebas:
```bash
make test
```

### Cálculo de Cobertura:
```bash
make test-coverage
```
Si desea examinar qué líneas de código específicas están cubiertas por las pruebas, puede generar y visualizar el informe HTML con:
```bash
go tool cover -html=coverage.out
```
