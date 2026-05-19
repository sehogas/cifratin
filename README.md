# Cifratin 🔐

[![Go Version](https://img.shields.io/github/go-mod/go-version/sehogas/cifratin?color=00ADD8&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/sehogas/cifratin)](https://goreportcard.com/report/github.com/sehogas/cifratin)
[![Security Policy](https://img.shields.io/badge/Security-Policy_Enabled-success?logo=github-security)](SECURITY.md)
[![Latest Release](https://img.shields.io/github/v/release/sehogas/cifratin?logo=github)](https://github.com/sehogas/cifratin/releases)

A command-line utility (CLI) written in Go designed to encrypt and decrypt files quickly and securely. It implements modern symmetric cryptography using the **AES-256-GCM** standard and provides a clean interface for processing individual files, search patterns (globbing), or entire directories recursively.

---

## 🛠️ Scope and Features

The project has been designed under the principles of simplicity and robustness, ensuring that confidential data management is performed securely at a local level.

- **Robust Cryptography**:
  - **Symmetric Encryption**: Uses **AES-256** in **GCM** (Galois/Counter Mode). This mode provides authenticated encryption, guaranteeing both **confidentiality** and **integrity** of the encrypted data (detecting if the file has been altered).
  - **Key Derivation**: User passwords are processed using **SHA-256** to consistently derive a 32-byte (256-bit) symmetric key, regardless of the original password's length.
- **Secure Password Input**: The security key is read interactively through the terminal using `golang.org/x/term`, hiding entered characters (no echo on screen) to prevent shoulder surfing attacks.
- **Input Flexibility**:
  - **Single File**: Encrypt or decrypt an individual file by specifying its paths.
  - **Search Patterns (Wildcards)**: Allows using expressions like `*.pdf` or `tests/*.txt` to batch process multiple matching files.
  - **Directories (Recursive Processing)**: Traverses entire folder structures and processes each file individually, replicating the directory structure at the target destination.
- **Integrated gRPC Service**: Includes a standalone gRPC server (`cmd/cifratin-server`) separate from the CLI to allow integration and execution of the AES encryption engine from remote clients or microservices.
- **Smart Extension Handling**:
  - Automatically appends the `.enc` extension when encrypting files if they do not already have it.
  - Removes the `.enc` extension when decrypting if it is present in the source file name.
- **Directory Creation**: If the destination output directory does not exist on the system, the application automatically creates it before proceeding with the file write operations.

---

## 🏗️ Prerequisites

To compile and run Cifratin on your local machine, you will need to have installed:

- **Go**: Version `1.25.1` or higher (as defined in the module [go.mod](go.mod)).
- **Make**: Automation tool (optional, but highly recommended to run project tasks easily).

> [!NOTE]
> **Modifications to the gRPC definition**: If you plan to modify the `.proto` files (inside `api/proto/v1/`), you will need the Protocol Buffers compiler (`protoc`) installed on your system along with the Go code generation plugins:
> - `protoc-gen-go` (installable via `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`)
> - `protoc-gen-go-grpc` (installable via `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`)
>
> The already generated Go files (`.pb.go`) are versioned in the repository, so installing these tools is not mandatory unless you modify the protocol API.

---

## 🚀 Installation and Compilation

To compile the project and generate the executable binaries, follow these steps:

```bash
# Compile the CLI executable binary
make build

# Compile the gRPC client binary
make build-client

# Compile the gRPC server binary
make build-server
```
The resulting CLI binary will be saved in `bin/cifratin`, the client in `bin/cifratin-client`, and the server in `bin/cifratin-server` (or with `.exe` in Windows environments).

### Using Go directly
```bash
go build -o bin/cifratin ./cmd/cifratin
go build -o bin/cifratin-client ./cmd/cifratin-client
go build -o bin/cifratin-server ./cmd/cifratin-server
```

If you want to clean up built files and test results:
```bash
make clean
```

---

## 📖 Usage Instructions

### 1. Local Encryptor (Traditional CLI)

The local CLI application accepts flags to parameterize its behavior:

```bash
bin/cifratin -mode=<encrypt|decrypt> -in=<source> -out=<destination>
```

---

### 2. Remote Encryptor (gRPC Client)

The gRPC client application replicates the same functionality as the local CLI but performs cryptographic tasks by calling the gRPC server remotely.

```bash
bin/cifratin-client -mode=<encrypt|decrypt> -in=<source> -out=<destination> [additional flags]
```

#### Additional Flags for the gRPC Client:
- `-addr`: IP address and port of the gRPC server. Default is `localhost:50051`.
- `-apikey`: API key to authenticate with the server. If not specified, it will attempt to read from the `CIFRATIN_API_KEY` environment variable, otherwise it will use the default key `"dev-key-123"`.

> [!IMPORTANT]
> For both the local and gRPC client versions, you will be prompted to enter the cryptographic password for the file interactively in the console.

---

### Practical Execution Examples

#### 1. Encrypt and Decrypt an Individual File (Local)

```bash
# Encrypt an individual PDF file
bin/cifratin -mode=encrypt -in=test.pdf -out=test.pdf.enc

# Decrypt the previously encrypted file
bin/cifratin -mode=decrypt -in=test.pdf.enc -out=test_restored.pdf
```

#### 2. Encrypt and Decrypt an Individual File (gRPC Client)

Ensure the gRPC server is running (`make run-server`).

```bash
# Encrypt an individual PDF file by sending it to the gRPC server
bin/cifratin-client -mode=encrypt -in=test.pdf -out=test.pdf.enc

# Decrypt the previously encrypted file via gRPC
bin/cifratin-client -mode=decrypt -in=test.pdf.enc -out=test_restored_grpc.pdf
```

#### 3. Encrypt and Decrypt a Complete Directory (gRPC Client)

The directory mode in the gRPC client also preserves the internal recursive folder structure.

```bash
# Encrypt all contents of the 'tests' folder to 'encrypted_output_grpc'
bin/cifratin-client -mode=encrypt -in=tests -out=encrypted_output_grpc

# Decrypt the encrypted content to 'restored_output_grpc'
bin/cifratin-client -mode=decrypt -in=encrypted_output_grpc -out=restored_output_grpc
```

---

## 🔒 Security in the gRPC Service

By default, exposing a gRPC service on a network can be vulnerable if you do not restrict who is allowed to invoke it. To mitigate this, the service implements an API-key-based authentication layer using a unary server interceptor (`UnaryServerInterceptor`).

### 1. Server Configuration
Authorized keys are defined on the server using the `CIFRATIN_API_KEYS` environment variable, separated by commas:

```bash
# On Windows systems (PowerShell)
$env:CIFRATIN_API_KEYS="my-service-a,my-service-b,api-gateway-key"
make run-server

# On Linux/macOS systems
export CIFRATIN_API_KEYS="my-service-a,my-service-b,api-gateway-key"
make run-server
```

> [!NOTE]
> If you start the server without setting this variable, the system will use a default key (`dev-key-123`) for development environments and issue a warning in the console.

### 2. Client Consumption
For a client service to successfully invoke the `EncryptFile` and `DecryptFile` endpoints, it must attach a valid key to the metadata of the gRPC call. The interceptor validates either of the following headers:

- **`x-api-key` header**: Must contain the raw token (e.g., `x-api-key: my-service-a`).
- **`authorization` header**: Supports standard Bearer tokens (e.g., `authorization: Bearer my-service-a`) or the raw token value.

Any call that lacks the header or provides an invalid key will be rejected at the interceptor level, returning the gRPC error codes `Unauthenticated` or `PermissionDenied` respectively.

---

## 🛠️ Development Automation (Makefile)

The project includes a `Makefile` with predefined commands to streamline common development and demonstration tasks:

| Command | Description |
| :--- | :--- |
| `make help` | Displays the Makefile help menu with all available options. |
| `make build` | Compiles the CLI code and generates the binary in the `bin/` folder. |
| `make build-server`| Compiles the gRPC server and generates its executable in `bin/`. |
| `make clean` | Removes compiled binaries, coverage files, and temporary outputs. |
| `make test` | Runs unit and integration tests, displaying the output in the console. |
| `make test-coverage` | Runs tests and calculates code coverage. |
| `make run-cifrar-archivo` | Runs an encryption demonstration using the `test.pdf` file. |
| `make run-descifrar-archivo`| Runs a decryption demonstration of the file encrypted in the previous task. |
| `make run-cifrar-carpeta` | Practical demonstration of recursive folder encryption of `tests/`. |
| `make run-descifrar-carpeta`| Decrypts the recursively processed folder back into plain text. |
| `make run-cifrar-patron` | Practical demonstration of encryption using file search patterns. |
| `make run-server` | Starts the gRPC server on port 50051 (enabled with reflection). |
| `make help-app` | Shows the native CLI usage instructions printed by Go. |

---

## 🧪 Testing and Code Coverage

The project contains a comprehensive suite of unit and integration tests covering the main success and error flows, such as:
- Automatic target file name adjustments.
- Successful encryption of individual data.
- Correct decryption using the original password.
- Handling failures due to incorrect keys (AES-GCM integrity checks).
- Handling corrupt or incomplete files.
- Behavior with invalid parameters or non-existent paths.

### Running Tests:
```bash
make test
```

### Coverage Calculation:
```bash
make test-coverage
```
If you wish to examine which specific lines of code are covered by the tests, you can generate and view the HTML report with:
```bash
go tool cover -html=coverage.out
```
