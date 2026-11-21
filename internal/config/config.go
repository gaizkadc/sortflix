package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config contains the settings for the sortflix execution.
type Config struct {
	ConfigFile        string   `json:"-"`
	Path              string   `json:"path"`
	MinSizeMB         int64    `json:"min_size_mb"`
	VideoExtensions   []string `json:"video_extensions"`
	DeleteJunk        bool     `json:"delete_junk"`
	CleanDirs         bool     `json:"clean_empty_dirs"`
	IgnoreArticles    bool     `json:"ignore_articles"`
	IgnoredArticles   []string `json:"ignored_articles"`
	IgnoredFiles      []string `json:"ignored_files"`
	CollisionStrategy string   `json:"collision_strategy"`
	DryRun            bool     `json:"dry_run"`
	Yes               bool     `json:"yes"`
	Verbose           bool     `json:"verbose"`
}

// DefaultVideoExtensions returns the default recognized video file extensions.
func DefaultVideoExtensions() []string {
	return []string{
		".mp4", ".mkv", ".avi", ".mov", ".m4v",
		".wmv", ".flv", ".webm", ".ts", ".m2ts",
		".vob", ".mpg", ".mpeg", ".iso", ".divx",
		".rmvb", ".asf",
	}
}

// DefaultIgnoredArticles returns common definite/indefinite articles across languages.
func DefaultIgnoredArticles() []string {
	return []string{
		"The", "A", "An", "El", "La", "Los", "Las", "Le", "Les", "Der", "Die", "Das",
	}
}

// DefaultIgnoredFiles returns files that should never be deleted as junk.
func DefaultIgnoredFiles() []string {
	return []string{
		".nomedia", ".driveignore", ".git",
	}
}

// DefaultConfig returns a Config instance with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MinSizeMB:         100,
		VideoExtensions:   DefaultVideoExtensions(),
		DeleteJunk:        true,
		CleanDirs:         true,
		IgnoreArticles:    false,
		IgnoredArticles:   DefaultIgnoredArticles(),
		IgnoredFiles:      DefaultIgnoredFiles(),
		CollisionStrategy: "rename",
		DryRun:            false,
		Yes:               false,
		Verbose:           false,
	}
}

// FindDefaultConfigFile looks for a config file in known default locations:
// 1. ./config.json
// 2. ./sortflix.json
// 3. ~/.config/sortflix/config.json
// 4. ~/.sortflix.json
func FindDefaultConfigFile() string {
	candidates := []string{
		"config.json",
		"sortflix.json",
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "sortflix", "config.json"))
		candidates = append(candidates, filepath.Join(home, ".sortflix.json"))
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}

	return ""
}

// LoadConfigFile loads and merges JSON configuration into cfg.
func LoadConfigFile(cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading config file '%s': %w", filePath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("error parsing JSON in config file '%s': %w", filePath, err)
	}

	cfg.ConfigFile = filePath
	return nil
}

// ParseFlags parses command-line arguments, loading a config file if available,
// and allowing CLI flags to override configured values.
func ParseFlags() (*Config, error) {
	cfg := DefaultConfig()

	// Check CLI args for an explicit -config / -c flag before general flag parsing
	var explicitConfigFile string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-config" || arg == "--config" || arg == "-c" {
			if i+1 < len(os.Args) {
				explicitConfigFile = os.Args[i+1]
			}
			break
		} else if strings.HasPrefix(arg, "-config=") || strings.HasPrefix(arg, "--config=") {
			explicitConfigFile = strings.TrimPrefix(strings.TrimPrefix(arg, "--config="), "-config=")
			break
		} else if strings.HasPrefix(arg, "-c=") {
			explicitConfigFile = strings.TrimPrefix(arg, "-c=")
			break
		}
	}

	// Determine which config file to load
	configToLoad := explicitConfigFile
	if configToLoad == "" {
		configToLoad = FindDefaultConfigFile()
	}

	if configToLoad != "" {
		if err := LoadConfigFile(cfg, configToLoad); err != nil {
			if explicitConfigFile != "" {
				return nil, err
			}
			// If it was a default discovery that failed to parse, report it
			return nil, fmt.Errorf("found default config '%s' but failed to load it: %w", configToLoad, err)
		}
	}

	// Temporary CLI flag variables initialized with current cfg values
	var (
		flagConfig         string
		flagPath           string
		flagMinSizeMB      int64
		flagDryRun         bool
		flagYes            bool
		flagDeleteJunk     bool
		flagCleanDirs      bool
		flagIgnoreArticles bool
		flagVerbose        bool
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	fs.StringVar(&flagConfig, "config", cfg.ConfigFile, "Path to JSON configuration file")
	fs.StringVar(&flagConfig, "c", cfg.ConfigFile, "Alias for -config")

	fs.StringVar(&flagPath, "path", cfg.Path, "Path to movies root directory (e.g. /Volumes/GoogleDrive/Movies)")
	fs.StringVar(&flagPath, "p", cfg.Path, "Alias for -path")

	fs.Int64Var(&flagMinSizeMB, "min-size", cfg.MinSizeMB, "Minimum size in MB for a video file to be considered a movie")

	fs.BoolVar(&flagDryRun, "dry-run", cfg.DryRun, "Dry-run mode: show what would be done without modifying any file")
	fs.BoolVar(&flagDryRun, "n", cfg.DryRun, "Alias for -dry-run")

	fs.BoolVar(&flagYes, "yes", cfg.Yes, "Skip confirmation prompt before applying changes")
	fs.BoolVar(&flagYes, "y", cfg.Yes, "Alias for -yes")

	fs.BoolVar(&flagDeleteJunk, "delete-junk", cfg.DeleteJunk, "Delete non-movie files (srt, txt, jpg, nfo, etc.)")
	fs.BoolVar(&flagCleanDirs, "clean-dirs", cfg.CleanDirs, "Delete empty subdirectories after moving movies")
	fs.BoolVar(&flagIgnoreArticles, "ignore-articles", cfg.IgnoreArticles, "Ignore leading articles like 'The', 'El', 'A' when categorizing")

	fs.BoolVar(&flagVerbose, "verbose", cfg.Verbose, "Show detailed logging during execution")
	fs.BoolVar(&flagVerbose, "v", cfg.Verbose, "Alias for -verbose")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sortflix - Movie library organizer & cleaner\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  sortflix [options]\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /path/to/movies [options]\n")
		fmt.Fprintf(os.Stderr, "  sortflix /path/to/movies [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  sortflix --dry-run\n")
		fmt.Fprintf(os.Stderr, "  sortflix -config config.json --dry-run\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /Volumes/GoogleDrive/Movies --min-size 150 -y\n")
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// Override config with flags that were explicitly passed on CLI
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "config", "c":
			cfg.ConfigFile = flagConfig
		case "path", "p":
			cfg.Path = flagPath
		case "min-size":
			cfg.MinSizeMB = flagMinSizeMB
		case "dry-run", "n":
			cfg.DryRun = flagDryRun
		case "yes", "y":
			cfg.Yes = flagYes
		case "delete-junk":
			cfg.DeleteJunk = flagDeleteJunk
		case "clean-dirs":
			cfg.CleanDirs = flagCleanDirs
		case "ignore-articles":
			cfg.IgnoreArticles = flagIgnoreArticles
		case "verbose", "v":
			cfg.Verbose = flagVerbose
		}
	})

	// If path was passed as positional argument, it takes highest precedence
	if fs.NArg() > 0 {
		cfg.Path = fs.Arg(0)
	}

	if strings.TrimSpace(cfg.Path) == "" {
		return nil, fmt.Errorf("you must specify the movies directory path in config file, via -path, or as a positional argument")
	}

	cfg.Path = ExpandPath(cfg.Path)

	info, err := os.Stat(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot access directory '%s': %w", cfg.Path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", cfg.Path)
	}

	if cfg.MinSizeMB < 0 {
		return nil, fmt.Errorf("minimum size cannot be negative")
	}

	return cfg, nil
}

// ExpandPath expands leading ~ to the user's home directory and makes the path absolute.
func ExpandPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "~/") || trimmed == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			if trimmed == "~" {
				trimmed = home
			} else {
				trimmed = filepath.Join(home, trimmed[2:])
			}
		}
	}

	if abs, err := filepath.Abs(trimmed); err == nil {
		return abs
	}
	return trimmed
}
