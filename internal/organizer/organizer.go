package organizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sortflix/internal/config"
)

// MoveAction represents a movie file to be relocated to its category folder.
type MoveAction struct {
	Source      string
	Destination string
	Size        int64
	Category    string
}

// DeleteFileAction represents a junk file to be removed.
type DeleteFileAction struct {
	Path string
	Size int64
}

// Plan represents all changes identified after scanning the movie library.
type Plan struct {
	RootDir          string
	MoviesFound      int
	MoviesInPlace    int
	Moves            []MoveAction
	FileDeletions    []DeleteFileAction
	DirDeletions     []string
	JunkBytesTotal   int64
	MovieBytesToMove int64
}

// Organizer orchestrates scanning, planning, and executing the library cleanup.
type Organizer struct {
	cfg          *config.Config
	videoExts    map[string]bool
	ignoredFiles map[string]bool
	catOpts      CategoryOptions
}

// New creates a new Organizer instance with the given configuration.
func New(cfg *config.Config) *Organizer {
	org := &Organizer{
		cfg:          cfg,
		videoExts:    make(map[string]bool),
		ignoredFiles: make(map[string]bool),
		catOpts: CategoryOptions{
			IgnoreArticles:  cfg.IgnoreArticles,
			IgnoredArticles: cfg.IgnoredArticles,
		},
	}

	// Index configured video extensions
	for _, ext := range cfg.VideoExtensions {
		clean := strings.ToLower(strings.TrimSpace(ext))
		if clean == "" {
			continue
		}
		if !strings.HasPrefix(clean, ".") {
			clean = "." + clean
		}
		org.videoExts[clean] = true
	}

	// Index ignored files (case-insensitive)
	for _, file := range cfg.IgnoredFiles {
		clean := strings.ToLower(strings.TrimSpace(file))
		if clean != "" {
			org.ignoredFiles[clean] = true
		}
	}

	return org
}

// isVideoFile checks if the path has one of the configured video extensions.
func (o *Organizer) isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return o.videoExts[ext]
}

// isIgnoredFile checks if the filename is explicitly protected from deletion (e.g. .driveignore, .nomedia).
func (o *Organizer) isIgnoredFile(name string) bool {
	base := strings.ToLower(name)
	return o.ignoredFiles[base]
}

// BuildPlan scans the directory tree and builds an execution plan without modifying files.
func (o *Organizer) BuildPlan() (*Plan, error) {
	plan := &Plan{
		RootDir:       o.cfg.Path,
		Moves:         make([]MoveAction, 0),
		FileDeletions: make([]DeleteFileAction, 0),
		DirDeletions:  make([]string, 0),
	}

	minSizeBytes := o.cfg.MinSizeMB * 1024 * 1024
	reservedDests := make(map[string]bool)

	// Collect existing category directories under root
	categoryDirs := o.detectCategoryDirs()

	// Track files per directory to simulate empty directories in dry-run/plan
	dirFilesRemaining := make(map[string]int)
	candidateDirs := make([]string, 0)

	err := filepath.WalkDir(o.cfg.Path, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if o.cfg.Verbose {
				fmt.Printf("[WARNING] Could not access '%s': %v\n", path, walkErr)
			}
			return nil
		}

		if path == o.cfg.Path {
			return nil
		}

		if d.IsDir() {
			// Do not consider top-level category dirs for deletion
			if !o.isTopLevelCategoryDir(path, categoryDirs) {
				candidateDirs = append(candidateDirs, path)
				if _, exists := dirFilesRemaining[path]; !exists {
					dirFilesRemaining[path] = 0
				}
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if o.cfg.Verbose {
				fmt.Printf("[WARNING] Error reading info for '%s': %v\n", path, err)
			}
			return nil
		}

		parentDir := filepath.Dir(path)
		dirFilesRemaining[parentDir]++

		isMovie := o.isVideoFile(path) && info.Size() >= minSizeBytes

		if isMovie {
			plan.MoviesFound++
			cat := GetCategoryWithOptions(filepath.Base(path), o.catOpts)
			catDirName := o.resolveCategoryDirName(cat, categoryDirs)
			targetDir := filepath.Join(o.cfg.Path, catDirName)
			targetFile := filepath.Join(targetDir, filepath.Base(path))

			// Check if already in target location
			if path == targetFile {
				plan.MoviesInPlace++
				return nil
			}

			// Handle destination collision
			finalDest := targetFile
			if reservedDests[finalDest] || o.destinationExists(finalDest, path) {
				if o.cfg.CollisionStrategy == "skip" {
					if o.cfg.Verbose {
						fmt.Printf("[SKIP COLLISION] Destination exists: %s\n", targetFile)
					}
					return nil
				}
				finalDest = o.resolvePlanCollision(targetFile, reservedDests)
			}
			reservedDests[finalDest] = true

			plan.Moves = append(plan.Moves, MoveAction{
				Source:      path,
				Destination: finalDest,
				Size:        info.Size(),
				Category:    catDirName,
			})
			plan.MovieBytesToMove += info.Size()

			// This file will move out of parentDir
			dirFilesRemaining[parentDir]--
		} else {
			// Non-movie file: check if it's protected or junk
			if o.isIgnoredFile(filepath.Base(path)) {
				// Don't mark as junk, but note it stays in directory
				return nil
			}

			if o.cfg.DeleteJunk {
				plan.FileDeletions = append(plan.FileDeletions, DeleteFileAction{
					Path: path,
					Size: info.Size(),
				})
				plan.JunkBytesTotal += info.Size()
				// This file will be deleted from parentDir
				dirFilesRemaining[parentDir]--
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking library: %w", err)
	}

	// Calculate which candidate directories will end up empty (bottom-up)
	if o.cfg.CleanDirs {
		plan.DirDeletions = o.simulateEmptyDirs(candidateDirs, dirFilesRemaining)
	}

	return plan, nil
}

// Execute applies the plan to the filesystem.
func (o *Organizer) Execute(plan *Plan) error {
	// Step 1: Execute movie moves
	movesDone := 0
	if len(plan.Moves) > 0 {
		fmt.Printf("  📦 Relocating %d movies to category folders...\n", len(plan.Moves))
		for i, move := range plan.Moves {
			if o.cfg.Verbose {
				fmt.Printf("     [%d/%d] %s -> %s (%s)\n",
					i+1, len(plan.Moves),
					o.relPath(move.Source),
					o.relPath(move.Destination),
					FormatBytes(move.Size))
			}
			if err := MoveFile(move.Source, move.Destination); err != nil {
				return fmt.Errorf("failed to move '%s' to '%s': %w", move.Source, move.Destination, err)
			}
			// Verify file exists at destination
			if _, err := os.Lstat(move.Destination); err != nil {
				return fmt.Errorf("move verification failed for '%s': destination does not exist after move: %w", move.Destination, err)
			}
			movesDone++
		}
		fmt.Printf("     ✓ %d movies successfully relocated.\n", movesDone)
	}

	// Step 2: Delete junk files
	filesDeleted := 0
	if o.cfg.DeleteJunk && len(plan.FileDeletions) > 0 {
		fmt.Printf("  🗑️  Deleting %d junk files (%s)...\n", len(plan.FileDeletions), FormatBytes(plan.JunkBytesTotal))
		for i, del := range plan.FileDeletions {
			if o.cfg.Verbose {
				fmt.Printf("     [%d/%d] %s (%s)\n", i+1, len(plan.FileDeletions), o.relPath(del.Path), FormatBytes(del.Size))
			}
			if err := RemoveFile(del.Path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete file '%s': %w", del.Path, err)
			}
			filesDeleted++
		}
		fmt.Printf("     ✓ %d junk files successfully deleted.\n", filesDeleted)
	}

	// Step 3: Remove empty subdirectories (bottom-up: longest path first)
	dirsRemoved := 0
	dirsSkipped := 0
	if o.cfg.CleanDirs && len(plan.DirDeletions) > 0 {
		fmt.Printf("  📂 Removing %d empty subdirectories...\n", len(plan.DirDeletions))
		sort.Slice(plan.DirDeletions, func(i, j int) bool {
			return len(plan.DirDeletions[i]) > len(plan.DirDeletions[j])
		})

		for _, dir := range plan.DirDeletions {
			// Clean any leftover system/metadata files (.DS_Store, Thumbs.db, ._* AppleDouble) before checking
			cleanSystemFilesInDir(dir)

			empty, err := IsDirEmpty(dir)
			if err != nil {
				dirsSkipped++
				continue
			}
			if empty {
				if err := RemoveDir(dir); err == nil {
					if o.cfg.Verbose {
						fmt.Printf("     [REMOVED DIR] %s\n", o.relPath(dir))
					}
					dirsRemoved++
				} else {
					if o.cfg.Verbose {
						fmt.Printf("     [WARN] Could not remove directory %s: %v\n", o.relPath(dir), err)
					}
					dirsSkipped++
				}
			} else {
				dirsSkipped++
			}
		}
		fmt.Printf("     ✓ %d empty directories removed (%d kept or non-empty).\n", dirsRemoved, dirsSkipped)
	}

	return nil
}

// detectCategoryDirs inspects rootDir for existing category directories (case-insensitive).
func (o *Organizer) detectCategoryDirs() map[string]string {
	result := make(map[string]string)
	entries, err := os.ReadDir(o.cfg.Path)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if entry.IsDir() && IsCategoryDir(entry.Name()) {
			lower := strings.ToLower(entry.Name())
			result[lower] = entry.Name()
		}
	}

	return result
}

// isTopLevelCategoryDir returns true if path is a category directory directly under rootDir.
func (o *Organizer) isTopLevelCategoryDir(path string, categoryDirs map[string]string) bool {
	if filepath.Dir(path) != o.cfg.Path {
		return false
	}
	baseLower := strings.ToLower(filepath.Base(path))
	_, exists := categoryDirs[baseLower]
	return exists || IsCategoryDir(filepath.Base(path))
}

// resolveCategoryDirName returns the actual directory name to use for a category.
// If a directory already exists (e.g. "A" instead of "a"), it reuses it. Otherwise defaults to lowercase.
func (o *Organizer) resolveCategoryDirName(cat string, existing map[string]string) string {
	lower := strings.ToLower(cat)
	if actual, found := existing[lower]; found {
		return actual
	}
	return lower
}

// destinationExists checks if target exists and is not the same as source.
func (o *Organizer) destinationExists(target, source string) bool {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}

	srcInfo, err := os.Lstat(source)
	if err != nil {
		return true
	}

	return !os.SameFile(info, srcInfo)
}

// resolvePlanCollision finds a non-colliding destination path considering both disk and reserved paths.
func (o *Organizer) resolvePlanCollision(target string, reserved map[string]bool) string {
	ext := filepath.Ext(target)
	dir := filepath.Dir(target)
	base := strings.TrimSuffix(filepath.Base(target), ext)

	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if !reserved[candidate] {
			if _, err := os.Lstat(candidate); os.IsNotExist(err) {
				return candidate
			}
		}
	}
}

// simulateEmptyDirs simulates bottom-up directory cleanup.
func (o *Organizer) simulateEmptyDirs(candidateDirs []string, fileCounts map[string]int) []string {
	// Sort by depth descending (deepest directories first)
	sort.Slice(candidateDirs, func(i, j int) bool {
		return len(candidateDirs[i]) > len(candidateDirs[j])
	})

	emptyDirs := make([]string, 0)
	deletedDirs := make(map[string]bool)

	for _, dir := range candidateDirs {
		// A directory is empty if its remaining file count is 0 and all its child directories are deleted
		if fileCounts[dir] == 0 {
			emptyDirs = append(emptyDirs, dir)
			deletedDirs[dir] = true
		}
	}

	return emptyDirs
}

func (o *Organizer) relPath(path string) string {
	rel, err := filepath.Rel(o.cfg.Path, path)
	if err != nil {
		return path
	}
	return rel
}

func cleanSystemFilesInDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if IsSystemFile(e.Name()) {
			_ = RemoveFile(filepath.Join(dir, e.Name()))
		}
	}
}

// FormatBytes formats byte counts into human-readable strings (B, KB, MB, GB).
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
