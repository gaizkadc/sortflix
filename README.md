# 🎬 Sortflix

Fast, dependency-free movie library organizer and cleaner written in Go.

Sortflix traverses your movies folder, flattens unnecessary nested subfolders, moves every movie (video files $\ge$ 100 MB) directly into its corresponding alphabetical directory (`#`, `a`, `b`, ..., `z`), deletes junk files (loose `.srt`, `.txt`, `.nfo`, scene stills, trailers/samples < 100 MB), and removes all emptied subdirectories.

The result is a clean, shallow, and predictable library structure:
```text
movies/
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

## 🚀 Key Features

- **Configuration File Support (`config.json`)**:
  - Automatically loads settings from `config.json`, `sortflix.json`, or `~/.config/sortflix/config.json`.
  - Configurable video extensions list, minimum file size, ignored files, and custom article stripping.
- **Smart Alphabetical Categorization**:
  - Titles starting with digits or symbols go to `#`.
  - Titles starting with letters go to `a` through `z`.
  - Normalizes accented and non-ASCII Latin characters (e.g., `Ágora` ➜ `a`, `Élite` ➜ `e`, `Ñu` ➜ `n`).
  - Automatically strips release tags and bracketed prefixes before categorization (e.g., `[1080p] The Matrix` ➜ `t`, `(2001) A Space Odyssey` ➜ `a`).
  - Optional article stripping (`ignore_articles: true`): `"The Matrix"` ➜ `m`, `"El Padrino"` ➜ `p`.
  - Preserves existing directory casing (e.g., reuses `A` if it already exists).
- **Size-based Video Filtering**:
  - Treats files as movies only if they match configured video extensions (`.mp4`, `.mkv`, `.avi`, `.mov`, `.iso`, etc.) and are $\ge$ 100 MB (customizable with `--min-size`).
- **Junk File Pruning**:
  - Automatically deletes non-movie artifacts (`.srt`, `.txt`, `.nfo`, `.jpg`, `.png`, trailers, and samples smaller than the threshold).
  - Protects files listed in `ignored_files` (e.g. `.driveignore`, `.nomedia`).
- **Directory Flattening & Empty Folder Pruning**:
  - Removes nested movie folders (e.g., `a/american ninja/` is removed once `american ninja.mp4` is moved up to `a/`).
  - **Master Directory Protection**: root `#` and `a`-`z` folders are never deleted.
- **Collision Protection**:
  - If two different movie files share the same filename in separate folders, Sortflix automatically renames them safely (e.g., `movie (1).mp4`) without overwriting.
- **Dry-Run Simulation (`--dry-run` / `-n`)**:
  - Preview every move, file deletion, and directory removal before touching disk.
- **Interactive Safety**:
  - Prompts for confirmation before applying changes (can be bypassed with `-y` / `--yes` in scripts).
- **Cross-Mount Safe**:
  - Uses atomic `os.Rename` with transparent fallback to copy + remove across distinct volumes or network mounts.

---

## ⚙️ Configuration File (`config.json`)

Sortflix automatically detects `config.json` in the current directory or `~/.config/sortflix/config.json`.
To create your own configuration file from the template, run:

```bash
make config
```

This will copy `config.example.json` to `config.json` (which is git-ignored so your private paths are never committed):

```json
{
  "path": "/path/to/your/movies",
  "min_size_mb": 100,
  "video_extensions": [
    ".mp4",
    ".mkv",
    ".avi",
    ".mov",
    ".m4v",
    ".wmv",
    ".flv",
    ".webm",
    ".ts",
    ".m2ts",
    ".vob",
    ".mpg",
    ".mpeg",
    ".iso",
    ".divx",
    ".rmvb",
    ".asf"
  ],
  "delete_junk": true,
  "clean_empty_dirs": true,
  "ignore_articles": false,
  "ignored_articles": [
    "The", "A", "An", "El", "La", "Los", "Las", "Le", "Les", "Der", "Die", "Das"
  ],
  "ignored_files": [
    ".DS_Store", "Thumbs.db", ".nomedia", ".driveignore"
  ],
  "collision_strategy": "rename",
  "dry_run": false,
  "verbose": false
}
```

---

## 🛠️ Build & Installation

Requires Go 1.22+.

```bash
# Clone or navigate to repository
cd sortflix

# Run tests
make test

# Build binary into bin/sortflix
make build
```

---

## 📖 Usage

### 1. Dry-Run Simulation (Recommended first)
Inspect what would be moved or deleted without modifying anything (uses path from `config.json` by default):

```bash
./bin/sortflix --dry-run
```
or using short flags:
```bash
./bin/sortflix -n
```

### 2. Interactive Execution
```bash
./bin/sortflix
```
Sortflix will show the plan and prompt:
`Do you want to apply these changes to your library? [y/N]: `

### 3. Override Settings via CLI
CLI flags always take precedence over `config.json`:
```bash
./bin/sortflix -path "/Volumes/External/Movies" --min-size 150 --dry-run
```

---

## ⚙️ Command-Line Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `-config` | `-c` | `config.json` | Path to JSON configuration file |
| `-path` | `-p` | _(from config)_ | Path to root movies folder |
| `-min-size` | | `100` | Minimum file size in MB for video files |
| `-dry-run` | `-n` | `false` | Dry-run mode: show plan without making changes |
| `-yes` | `-y` | `false` | Apply changes without interactive confirmation |
| `-delete-junk` | | `true` | Delete non-movie files (srt, txt, jpg, nfo, etc.) |
| `-clean-dirs` | | `true` | Delete subdirectories that become empty |
| `-ignore-articles` | | `false` | Ignore leading articles (The, El, La, A, ...) |
| `-verbose` | `-v` | `false` | Display detailed log for every processed item |

---

## 🧪 Project Layout

```text
sortflix/
├── cmd/
│   └── sortflix/
│       └── main.go          # CLI entrypoint, flag parsing, formatted console UI
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration model, JSON loader & flag overrides
│   │   └── config_test.go   # Configuration unit tests
│   └── organizer/
│       ├── category.go      # Categorization, title normalization & article stripper
│       ├── category_test.go # Comprehensive categorization unit tests
│       ├── fsutil.go        # Video file checks, collision handler, file mover
│       ├── fsutil_test.go   # File system utility tests
│       ├── organizer.go     # Core scan, plan, and execution engine
│       └── organizer_test.go# Full end-to-end integration tests
├── config.json              # Default user configuration file
├── Makefile                 # Build, test, and clean automation
├── go.mod                   # Zero external dependencies (Go standard library)
└── README.md                # Documentation
```
