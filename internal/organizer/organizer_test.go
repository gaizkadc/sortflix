package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"sortflix/internal/config"
)

// helper to create a file of a specific size with dummy bytes
func createDummyFile(t *testing.T, path string, size int64) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
	defer f.Close()

	if size > 0 {
		if err := f.Truncate(size); err != nil {
			t.Fatalf("failed to set size for %s: %v", path, err)
		}
	}
}

func TestOrganizerEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()

	const minSizeMB = 1
	const movieSize = 2 * 1024 * 1024 // 2MB (qualifies as movie)
	const sampleSize = 500 * 1024      // 500KB (junk video)

	// Structure to test:
	// root/
	//   a/
	//     american ninja/
	//       american ninja.mp4 (movie)
	//       poster.jpg (junk)
	//       info.txt (junk)
	//       subs/
	//         es.srt (junk)
	//   b/
	//     batman.mkv (already in place!)
	//   c/
	//     sample.mp4 (too small: junk)
	//   #/
	//     300/
	//       300.mp4 (movie)
	//   empty_folder/

	createDummyFile(t, filepath.Join(tmpDir, "a", "american ninja", "american ninja.mp4"), movieSize)
	createDummyFile(t, filepath.Join(tmpDir, "a", "american ninja", "poster.jpg"), 1024)
	createDummyFile(t, filepath.Join(tmpDir, "a", "american ninja", "info.txt"), 512)
	createDummyFile(t, filepath.Join(tmpDir, "a", "american ninja", "subs", "es.srt"), 2048)
	createDummyFile(t, filepath.Join(tmpDir, "b", "batman.mkv"), movieSize)
	createDummyFile(t, filepath.Join(tmpDir, "c", "sample.mp4"), sampleSize)
	createDummyFile(t, filepath.Join(tmpDir, "#", "300", "300.mp4"), movieSize)

	emptyDir := filepath.Join(tmpDir, "empty_folder")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Path = tmpDir
	cfg.MinSizeMB = minSizeMB
	cfg.DeleteJunk = true
	cfg.CleanDirs = true
	cfg.DryRun = false
	cfg.Verbose = false

	org := New(cfg)

	// 1. Build Plan
	plan, err := org.BuildPlan()
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	if plan.MoviesFound != 3 {
		t.Errorf("expected 3 movies found, got %d", plan.MoviesFound)
	}
	if plan.MoviesInPlace != 1 {
		t.Errorf("expected 1 movie in place, got %d", plan.MoviesInPlace)
	}
	if len(plan.Moves) != 2 {
		t.Errorf("expected 2 moves, got %d", len(plan.Moves))
	}
	// Junk files: poster.jpg, info.txt, es.srt, sample.mp4 = 4
	if len(plan.FileDeletions) != 4 {
		t.Errorf("expected 4 file deletions, got %d", len(plan.FileDeletions))
	}

	// 2. Execute
	if err := org.Execute(plan); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify movies are in correct root category folders
	expectedMovies := []string{
		filepath.Join(tmpDir, "a", "american ninja.mp4"),
		filepath.Join(tmpDir, "b", "batman.mkv"),
		filepath.Join(tmpDir, "#", "300.mp4"),
	}
	for _, m := range expectedMovies {
		if _, err := os.Stat(m); os.IsNotExist(err) {
			t.Errorf("expected movie %s to exist", m)
		}
	}

	// Verify junk subdirectories are gone
	removedDirs := []string{
		filepath.Join(tmpDir, "a", "american ninja", "subs"),
		filepath.Join(tmpDir, "a", "american ninja"),
		filepath.Join(tmpDir, "#", "300"),
		filepath.Join(tmpDir, "empty_folder"),
	}
	for _, d := range removedDirs {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("expected directory %s to be deleted", d)
		}
	}

	// Verify category folders STILL exist
	categoryDirs := []string{
		filepath.Join(tmpDir, "a"),
		filepath.Join(tmpDir, "b"),
		filepath.Join(tmpDir, "c"),
		filepath.Join(tmpDir, "#"),
	}
	for _, cd := range categoryDirs {
		if _, err := os.Stat(cd); os.IsNotExist(err) {
			t.Errorf("category folder %s should NOT be deleted", cd)
		}
	}
}

func TestCollisionHandling(t *testing.T) {
	tmpDir := t.TempDir()

	const movieSize = 2 * 1024 * 1024
	// Two different subfolders with the same movie filename
	file1 := filepath.Join(tmpDir, "a", "rip1", "american ninja.mp4")
	file2 := filepath.Join(tmpDir, "a", "rip2", "american ninja.mp4")

	createDummyFile(t, file1, movieSize)
	createDummyFile(t, file2, movieSize)

	cfg := config.DefaultConfig()
	cfg.Path = tmpDir
	cfg.MinSizeMB = 1
	cfg.DeleteJunk = true
	cfg.CleanDirs = true
	cfg.DryRun = false
	cfg.Verbose = false

	org := New(cfg)
	plan, err := org.BuildPlan()
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	if len(plan.Moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(plan.Moves))
	}

	dest1 := plan.Moves[0].Destination
	dest2 := plan.Moves[1].Destination

	if dest1 == dest2 {
		t.Fatalf("destinations must be different to prevent overwriting: %s vs %s", dest1, dest2)
	}

	if err := org.Execute(plan); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, err := os.Stat(dest1); os.IsNotExist(err) {
		t.Errorf("dest1 %s does not exist", dest1)
	}
	if _, err := os.Stat(dest2); os.IsNotExist(err) {
		t.Errorf("dest2 %s does not exist", dest2)
	}
}

func TestAppleDoubleAndStrayFilesCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Folder structure like "The Godfather" with nested bonus folder and ._ files
	movieFolder := filepath.Join(tmpDir, "t", "The Godfather (1972)")
	extraFolder := filepath.Join(movieFolder, "Featurettes")

	// Movie moved out earlier or present
	movieFile := filepath.Join(movieFolder, "The Godfather (1972).mp4")
	createDummyFile(t, movieFile, 2*1024*1024)

	// 368-byte AppleDouble metadata files
	appleDouble1 := filepath.Join(movieFolder, "._The Godfather (1972).mp4")
	appleDouble2 := filepath.Join(movieFolder, "._metadata.xml")
	dsStore := filepath.Join(movieFolder, ".DS_Store")
	extraNfo := filepath.Join(extraFolder, "info.nfo")
	extraAppleDouble := filepath.Join(extraFolder, "._info.nfo")

	createDummyFile(t, appleDouble1, 368)
	createDummyFile(t, appleDouble2, 368)
	createDummyFile(t, dsStore, 6148)
	createDummyFile(t, extraNfo, 200)
	createDummyFile(t, extraAppleDouble, 368)

	cfg := config.DefaultConfig()
	cfg.Path = tmpDir
	cfg.MinSizeMB = 1
	cfg.DeleteJunk = true
	cfg.CleanDirs = true
	cfg.DryRun = false

	org := New(cfg)
	plan, err := org.BuildPlan()
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	if err := org.Execute(plan); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Movie should exist in t/
	expectedMovie := filepath.Join(tmpDir, "t", "The Godfather (1972).mp4")
	if _, err := os.Stat(expectedMovie); os.IsNotExist(err) {
		t.Errorf("expected movie %s to exist", expectedMovie)
	}

	// The Godfather folder and its subfolder should be completely gone!
	if _, err := os.Stat(movieFolder); !os.IsNotExist(err) {
		t.Errorf("expected folder %s to be completely removed, but it still exists", movieFolder)
	}
}
