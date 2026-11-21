# 🎬 Sortflix

Fast, dependency-free movie library organizer and cleaner for **Infuse** (Google Drive, macOS, NAS).

Sortflix traverses your movies folder, flattens unnecessary nested subfolders, moves every movie (video files $\ge$ 100 MB) to its corresponding alphabetical directory (`#`, `a`, `b`, ..., `z`), deletes junk files (loose `.srt`, `.txt`, `.nfo`, scene stills, trailers/samples < 100 MB), and deletes all emptied subdirectories.

The result is a clean, shallow structure optimized for Infuse browsing:
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

- **Smart Alphabetical Categorization**:
  - Titles starting with digits or symbols go to `#`.
  - Titles starting with letters go to `a` through `z`.
  - Normalizes accented and non-ASCII Latin characters (e.g., `Ágora` ➜ `a`, `Élite` ➜ `e`, `Ñu` ➜ `n`).
  - Automatically strips release tags and bracketed prefixes before categorization (e.g., `[1080p] The Matrix` ➜ `t`, `(2001) A Space Odyssey` ➜ `a`).
  - Preserves existing directory casing (e.g., reuses `A` if it already exists).
- **Size-based Video Filtering**:
  - Treats files as movies only if they have a recognized video extension (`.mkv`, `.mp4`, `.avi`, `.mov`, etc.) and are $\ge$ 100 MB (customizable with `--min-size`).
- **Junk File Pruning**:
  - Automatically deletes non-movie artifacts (`.srt`, `.txt`, `.nfo`, `.jpg`, `.png`, trailers, and samples smaller than the threshold).
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
Inspect what would be moved or deleted without modifying anything:

```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Movies" --dry-run
```
or using short flags:
```bash
./bin/sortflix "/Volumes/GoogleDrive/Movies" -n
```

### 2. Interactive Execution
```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Movies"
```
Sortflix will show the plan and prompt:
`Do you want to apply these changes to your library? [y/N]: `

### 3. Non-interactive / Scripted Execution
```bash
./bin/sortflix -path "/Volumes/GoogleDrive/Movies" -y
```

---

## ⚙️ Command-Line Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `-path` | `-p` | _(required)_ | Path to root movies folder |
| `-min-size` | | `100` | Minimum file size in MB for video files |
| `-dry-run` | `-n` | `false` | Dry-run mode: show plan without making changes |
| `-yes` | `-y` | `false` | Apply changes without interactive confirmation |
| `-delete-junk` | | `true` | Delete non-movie files (srt, txt, jpg, nfo, etc.) |
| `-clean-dirs` | | `true` | Delete subdirectories that become empty |
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
│   │   ├── config.go        # Configuration model & validation
│   │   └── config_test.go   # Configuration unit tests
│   └── organizer/
│       ├── category.go      # Categorization & title normalization engine (#, a-z)
│       ├── category_test.go # Comprehensive categorization unit tests
│       ├── fsutil.go        # Video file checks, collision handler, file mover
│       ├── fsutil_test.go   # File system utility tests
│       ├── organizer.go     # Core scan, plan, and execution engine
│       └── organizer_test.go# Full end-to-end integration tests
├── Makefile                 # Build, test, and clean automation
├── go.mod                   # Zero external dependencies
└── README.md                # Documentation
```
