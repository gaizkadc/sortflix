package organizer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Supported video file extensions (lowercase with leading dot).
var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".m4v":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".ts":   true,
	".vob":  true,
	".mpg":  true,
	".mpeg": true,
	".m2ts": true,
	".ogv":  true,
	".divx": true,
}

// IsVideoFile checks if the given file has a recognised video extension.
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExtensions[ext]
}

// IsSystemFile checks if the filename is a harmless system/metadata file like .DS_Store, Thumbs.db, or AppleDouble ._ files.
func IsSystemFile(filename string) bool {
	base := strings.ToLower(filepath.Base(filename))
	return base == ".ds_store" || base == "thumbs.db" || base == "desktop.ini" || strings.HasPrefix(base, "._")
}

// RemoveFile attempts to remove a file. If permission denied occurs (e.g. read-only rip file),
// it attempts to make it writable before retrying removal.
func RemoveFile(path string) error {
	err := os.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	if os.IsPermission(err) {
		if chmodErr := os.Chmod(path, 0666); chmodErr == nil {
			return os.Remove(path)
		}
	}
	return err
}

// RemoveDir attempts to remove an empty directory. If permission denied occurs,
// it attempts to make it writable before retrying.
func RemoveDir(dir string) error {
	err := os.Remove(dir)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	if os.IsPermission(err) {
		if chmodErr := os.Chmod(dir, 0777); chmodErr == nil {
			return os.Remove(dir)
		}
	}
	return err
}

// MoveFile moves a file from src to dst. It first attempts an atomic os.Rename.
// If cross-device or permission issues occur, it falls back to copy + remove.
func MoveFile(src, dst string) error {
	if src == dst {
		return nil
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed creating directory '%s': %w", dstDir, err)
	}

	// Try atomic rename
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback to copy + remove (useful across different mount points or network drives)
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed opening source file '%s': %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed reading source file info '%s': %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed creating destination file '%s': %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed copying from '%s' to '%s': %w", src, dst, err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed syncing destination file '%s': %w", dst, err)
	}

	srcFile.Close()
	dstFile.Close()

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed removing source file after copy '%s': %w", src, err)
	}

	return nil
}

// ResolveCollision generates an alternative path if dst already exists,
// appending " (1)", " (2)", etc. before the extension.
func ResolveCollision(dst string) string {
	if _, err := os.Lstat(dst); os.IsNotExist(err) {
		return dst
	}

	ext := filepath.Ext(dst)
	dir := filepath.Dir(dst)
	base := strings.TrimSuffix(filepath.Base(dst), ext)

	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Lstat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// IsDirEmpty checks if a directory contains any files or folders, ignoring harmless system files.
func IsDirEmpty(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !IsSystemFile(entry.Name()) {
			return false, nil
		}
	}

	return true, nil
}
