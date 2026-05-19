# 📦 Guía de Lanzamiento y Versionamiento de Cifratin

Este documento detalla el procedimiento técnico para definir versiones y publicar nuevos lanzamientos (releases) de los binarios ejecutables de Cifratin (`cifratin`, `cifratin-client`, `cifratin-server`, `cifratin-http-client` y `cifratin-http-server`) en el repositorio.

---

## 🏗️ Arquitectura de Versionamiento

Para evitar mantener valores estáticos (hardcodeados) en el código Go que deban editarse manualmente antes de compilar, el proyecto implementa **inyección dinámica de variables mediante el linker de Go** (`-ldflags`).

### 1. Variables de Versión en el Código
Cada punto de entrada de la aplicación define una variable global de paquete llamada `Version` inicializada por defecto con el valor `"dev"`:
- `cmd/cifratin/main.go`
- `cmd/cifratin-client/main.go`
- `cmd/cifratin-server/main.go`
- `cmd/cifratin-http-client/main.go`
- `cmd/cifratin-http-server/main.go`

### 2. Inyección local (Desarrollo)
Cuando se compila localmente de forma manual usando el `Makefile` (por ejemplo, `make build`, `make build-client`, `make build-server`, `make build-http-client` o `make build-http-server`), se utiliza el flag `-X` para sobreescribir el valor de la variable en tiempo de enlace:
```bash
go build -ldflags "-s -w -X main.Version=local-dev" -o bin/cifratin ./cmd/cifratin
```
Esto asegura que, durante el desarrollo local, al ejecutar `bin/cifratin -version` o `bin/cifratin-http-server -version` se muestre la leyenda `local-dev`.

### 3. Inyección en Producción (CI/CD)
Durante la ejecución del pipeline automatizado, **GoReleaser** inyecta de forma dinámica el número del Tag de Git correspondiente (ej. `v1.0.0`) reemplazando el valor `"dev"` de la variable `Version`.

---

## 🚀 Proceso para Publicar una Nueva Versión

El pipeline de entrega continua está automatizado usando **GoReleaser** y **GitHub Actions**. El flujo de compilación multiplataforma solo se activa cuando se empuja un nuevo **Tag de Git** con formato de versión semántica (ej: `v1.0.0`).

Siga estos pasos estructurados para publicar una versión:

### Paso 1: Validar el estado local
Antes de crear un lanzamiento, asegúrese de que la base de código compile correctamente y pase todas las pruebas automatizadas:
```bash
# Ejecutar la suite completa de pruebas unitarias y de integración
make test
```

### Paso 2: Crear el Tag de Git
Cree un tag anotado localmente indicando la versión que desea publicar. Use la convención de versionamiento semántico (`vX.Y.Z`):
```bash
git tag -a v1.0.0 -m "Release v1.0.0: descripción detallada del lanzamiento"
```

### Paso 3: Empujar los cambios y el Tag a GitHub
Suba los cambios acumulados de la rama principal junto con el nuevo tag al repositorio remoto:
```bash
# Empujar commits en la rama actual
git push origin main

# Empujar el tag creado
git push origin v1.0.0
```

---

## 🛠️ Automatización tras el Despliegue (CI/CD)

Una vez que GitHub recibe el tag que cumple con el patrón `v*`, se dispara automáticamente la acción configurada en `.github/workflows/release.yml`.

### ¿Qué hace el Pipeline automatizado?
1. **Configura el entorno de construcción:** Levanta un ejecutor virtual Linux e instala el SDK de Go estable.
2. **Construcción Multiplataforma paralela:** GoReleaser compila de forma nativa los tres ejecutables (`cifratin`, `cifratin-client` y `cifratin-server`) para los siguientes destinos:
   - **Windows** (`amd64` y `arm64`) en formato comprimido `.zip`.
   - **Linux** (`amd64` y `arm64`) en formato comprimido `.tar.gz`.
   - **macOS** (`amd64` y `arm64` Apple Silicon) en formato comprimido `.tar.gz`.
3. **Inyección de Metadatos:** Sobreescribe `main.Version` con el tag `v1.0.0`.
4. **Publicación del Release:** Crea la sección **Releases** en GitHub, escribe un changelog automático a partir de la historia de git y publica los paquetes compilados adjuntos listos para descarga.
