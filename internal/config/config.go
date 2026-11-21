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

	flag.StringVar(&cfg.Path, "path", "", "Path to movies root directory (e.g. /Volumes/GoogleDrive/Movies)")
	flag.StringVar(&cfg.Path, "p", "", "Alias for -path")
	flag.Int64Var(&cfg.MinSizeMB, "min-size", 100, "Minimum size in MB for a video file to be considered a movie")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Dry-run mode: show what would be done without modifying any file")
	flag.BoolVar(&cfg.DryRun, "n", false, "Alias for -dry-run")
	flag.BoolVar(&cfg.Yes, "yes", false, "Skip confirmation prompt before applying changes")
	flag.BoolVar(&cfg.Yes, "y", false, "Alias for -yes")
	flag.BoolVar(&cfg.DeleteJunk, "delete-junk", true, "Delete non-movie files (srt, txt, jpg, nfo, etc.)")
	flag.BoolVar(&cfg.CleanDirs, "clean-dirs", true, "Delete empty subdirectories after moving movies")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Show detailed logging during execution")
	flag.BoolVar(&cfg.Verbose, "v", false, "Alias for -verbose")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sortflix - Movie library organizer & cleaner for Infuse\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /path/to/movies [options]\n")
		fmt.Fprintf(os.Stderr, "  sortflix /path/to/movies [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /Volumes/GoogleDrive/Movies --dry-run\n")
		fmt.Fprintf(os.Stderr, "  sortflix -path /Volumes/GoogleDrive/Movies --min-size 150 -y\n")
	}

	flag.Parse()

	// If -path was not passed via flag, check positional argument
	if cfg.Path == "" && flag.NArg() > 0 {
		cfg.Path = flag.Arg(0)
	}

	if cfg.Path == "" {
		return nil, fmt.Errorf("you must specify the movies directory path with -path or as a positional argument")
	}

	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	cfg.Path = absPath

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
