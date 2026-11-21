# 🎬 Sortflix

Organizador y limpiador de biblioteca de películas para **Infuse** (Google Drive / macOS / NAS).

Sortflix recorre tu carpeta de películas, aplana subniveles innecesarios, mueve cada película (archivos de vídeo de más de 100 MB) a su carpeta alfabética correspondiente (`#`, `a`, `b`, ..., `z`), elimina archivos basura residuales (`.txt`, `.srt` sueltos, fotos, `.nfo`, vídeos de muestra/trailers < 100 MB) y suprime todas las subcarpetas que queden vacías.

El resultado es una estructura limpia e ideal para Infuse:
```text
pelis/
├── #/
│   └── 300.mp4
├── a/
│   └── american ninja.mp4
├── b/
│   └── batman begins.mkv
...
└── z/
```

---

## 🚀 Características principales

- **Clasificación alfabética inteligente**:
  - Películas que empiezan por dígitos van a `#`.
  - Películas de la `a` a la `z` a sus respectivas carpetas.
  - Normalización de acentos y diacríticos (ej: `Ágora` ➜ `a`, `Élite` ➜ `e`, `Ñu` ➜ `n`).
  - Detección y limpieza de prefijos típicos de rips (ej: `[1080p] The Matrix` ➜ `t`, `(2001) A Space Odyssey` ➜ `a`).
- **Filtrado por tamaño (umbral de vídeo)**:
  - Considera películas solo aquellos archivos con extensión de vídeo (`.mkv`, `.mp4`, `.avi`, `.mov`, etc.) con un tamaño $\ge$ 100 MB (configurable con `--min-size`).
- **Limpieza de archivos basura**:
  - Elimina automáticamente subtítulos no embebidos `.srt`, archivos `.txt`, `.nfo`, imágenes de carátulas o fotogramas `.jpg`/`.png` y vídeos menores al umbral (muestras, trailers o rips corruptos).
- **Eliminación de subcarpetas vacías**:
  - Borra todas las subcarpetas que hayan quedado vacías tras mover las películas (ej: `a/american ninja/` es suprimida).
  - **Protección de carpetas maestras**: `#` y `a`-`z` se preservan siempre.
- **Protección contra colisiones**:
  - Si dos películas en carpetas distintas tienen el mismo nombre, renombra de forma segura (ej: `american ninja (1).mp4`) para evitar sobrescrituras accidentales.
- **Modo Simulación (`--dry-run` / `-n`)**:
  - Muestra un informe exhaustivo con cada cambio que se realizaría antes de tocar un solo byte.
- **Confirmación interactiva**:
  - Pide confirmación al usuario antes de aplicar cambios reales (a menos que se pase `-y` / `--yes`).
- **Soporte entre volúmenes / Google Drive**:
  - Movimiento atómico (`os.Rename`) con fallback automático a copia y borrado si se opera entre sistemas de archivos o unidades de red.

---

## 🛠️ Instalación y Compilación

Requisitos: Go 1.22 o superior instalado.

```bash
# Clonar o entrar en el directorio del proyecto
cd sortflix

# Ejecutar tests unitarios
make test

# Compilar binario en bin/sortflix
make build
```

El binario compilado estará listo en `bin/sortflix`.

---

## 📖 Modo de Uso

### 1. Simulación previa (Recomendado)
Antes de mover o borrar nada, puedes comprobar qué se va a hacer:

```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Pelis" --dry-run
```
o de forma abreviada:
```bash
./bin/sortflix "/Volumes/GoogleDrive/Pelis" -n
```

### 2. Ejecutar con confirmación interactiva
```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Pelis"
```
El programa mostrará el resumen de movimientos y archivos a eliminar y te preguntará:
`¿Deseas aplicar estos cambios en tu biblioteca? [s/N]: `

### 3. Ejecutar automáticamente (sin preguntar)
Ideal para scripts o automatizaciones:
```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Pelis" -y
```

---

## ⚙️ Opciones y Parámetros

| Flag | Alias | Por defecto | Descripción |
| :--- | :--- | :--- | :--- |
| `-path` | `-p` | _(requerido)_ | Ruta a la carpeta raíz de películas |
| `-min-size` | | `100` | Tamaño mínimo en MB para considerar un vídeo como película |
| `-dry-run` | `-n` | `false` | Modo simulación: no modifica ningún archivo |
| `-yes` | `-y` | `false` | Aplica los cambios sin pedir confirmación interactiva |
| `-delete-junk` | | `true` | Elimina archivos no-película (srt, txt, jpg, etc.) |
| `-clean-dirs` | | `true` | Elimina subcarpetas que queden vacías tras mover vídeos |
| `-verbose` | `-v` | `false` | Muestra cada archivo y carpeta procesada en detalle |

---

## 🧪 Estructura del Proyecto

```text
sortflix/
├── cmd/
│   └── sortflix/
│       └── main.go          # Punto de entrada CLI, flags y reportes interactivos
├── internal/
│   ├── config/
│   │   ├── config.go        # Definición y validación de parámetros de configuración
│   │   └── config_test.go   # Tests de configuración
│   └── organizer/
│       ├── category.go      # Algoritmo de normalización y categorización (#, a-z)
│       ├── category_test.go # Tests de categorización y casos especiales
│       ├── fsutil.go        # Detección de formatos de vídeo, colisiones y utilidades FS
│       ├── fsutil_test.go   # Tests de utilidades de archivo
│       ├── organizer.go     # Motor de planificación y ejecución
│       └── organizer_test.go# Tests de integración extremo a extremo
├── Makefile                 # Automatización de build, test y clean
├── go.mod                   # Módulo de Go (cero dependencias externas)
└── README.md                # Documentación del proyecto
```
