package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config contains the settings for the sortflix execution.
type Config struct {
	Path        string // Root directory containing the movies library
	MinSizeMB   int64  // Minimum size in MB for a video file to be considered a movie
	DryRun      bool   // If true, simulate operations without altering the filesystem
	Yes         bool   // If true, skip confirmation prompt when executing changes
	DeleteJunk  bool   // If true, remove non-movie files (subtitles, images, nfo, txt, etc.)
	CleanDirs   bool   // If true, remove subdirectories that end up empty
	Verbose     bool   // If true, print detailed logging for every item processed
}

// ParseFlags parses command-line arguments into a Config instance.
func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Path, "path", "", "Ruta a la carpeta raíz de películas (ej: /Volumes/GoogleDrive/Pelis)")
	flag.StringVar(&cfg.Path, "p", "", "Alias de -path")
	flag.Int64Var(&cfg.MinSizeMB, "min-size", 100, "Tamaño mínimo en MB para considerar un archivo como película")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Modo simulación: muestra lo que se haría sin modificar ningún archivo")
	flag.BoolVar(&cfg.DryRun, "n", false, "Alias de -dry-run")
	flag.BoolVar(&cfg.Yes, "yes", false, "No pedir confirmación antes de aplicar cambios")
	flag.BoolVar(&cfg.Yes, "y", false, "Alias de -yes")
	flag.BoolVar(&cfg.DeleteJunk, "delete-junk", true, "Eliminar archivos no-película (srt, txt, jpg, nfo, etc.)")
	flag.BoolVar(&cfg.CleanDirs, "clean-dirs", true, "Eliminar subcarpetas vacías tras mover las películas")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Mostrar información detallada durante el proceso")
	flag.BoolVar(&cfg.Verbose, "v", false, "Alias de -verbose")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sortflix - Organizador y limpiador de biblioteca de películas para Infuse\n\n")
		fmt.Fprintf(os.Stderr, "Uso:\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /ruta/a/pelis [opciones]\n")
		fmt.Fprintf(os.Stderr, "  sortflix /ruta/a/pelis [opciones]\n\n")
		fmt.Fprintf(os.Stderr, "Opciones:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEjemplos:\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /Volumes/GoogleDrive/Pelis --dry-run\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /Volumes/GoogleDrive/Pelis --min-size 150 -y\n")
	}

	flag.Parse()

	// If -path was not passed via flag, check positional argument
	if cfg.Path == "" && flag.NArg() > 0 {
		cfg.Path = flag.Arg(0)
	}

	if cfg.Path == "" {
		return nil, fmt.Errorf("debes especificar la ruta a la carpeta de películas con -path o como argumento")
	}

	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("ruta inválida: %w", err)
	}
	cfg.Path = absPath

	info, err := os.Stat(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo acceder a la carpeta '%s': %w", cfg.Path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("la ruta '%s' no es un directorio", cfg.Path)
	}

	if cfg.MinSizeMB < 0 {
		return nil, fmt.Errorf("el tamaño mínimo no puede ser negativo")
	}

	return cfg, nil
}
