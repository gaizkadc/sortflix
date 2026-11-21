package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"movie.mp4", true},
		{"movie.MKV", true},
		{"movie.avi", true},
		{"movie.MOV", true},
		{"movie.m4v", true},
		{"movie.wmv", true},
		{"movie.flv", true},
		{"movie.webm", true},
		{"movie.ts", true},
		{"movie.srt", false},
		{"movie.txt", false},
		{"movie.nfo", false},
		{"movie.jpg", false},
		{"movie.png", false},
		{"movie", false},
		{".mp4", true},
	}

	for _, tt := range tests {
		got := IsVideoFile(tt.filename)
		if got != tt.expected {
			t.Errorf("IsVideoFile(%q) = %v, expected %v", tt.filename, got, tt.expected)
		}
	}
}

func TestIsSystemFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{".DS_Store", true},
		{"Thumbs.db", true},
		{"._movie.mp4", true},
		{"regular.txt", false},
		{"movie.mp4", false},
	}

	for _, tt := range tests {
		got := IsSystemFile(tt.filename)
		if got != tt.expected {
			t.Errorf("IsSystemFile(%q) = %v, expected %v", tt.filename, got, tt.expected)
		}
	}
}

func TestResolveCollision(t *testing.T) {
	tmpDir := t.TempDir()
	original := filepath.Join(tmpDir, "test.mp4")

	// Doesn't exist yet
	if got := ResolveCollision(original); got != original {
		t.Errorf("ResolveCollision for non-existent file should be %s, got %s", original, got)
	}

	// Create test.mp4
	if err := os.WriteFile(original, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFirstCollision := filepath.Join(tmpDir, "test (1).mp4")
	if got := ResolveCollision(original); got != expectedFirstCollision {
		t.Errorf("ResolveCollision for first collision should be %s, got %s", expectedFirstCollision, got)
	}

	// Create test (1).mp4
	if err := os.WriteFile(expectedFirstCollision, []byte("col1"), 0644); err != nil {
		t.Fatal(err)
	}

	expectedSecondCollision := filepath.Join(tmpDir, "test (2).mp4")
	if got := ResolveCollision(original); got != expectedSecondCollision {
		t.Errorf("ResolveCollision for second collision should be %s, got %s", expectedSecondCollision, got)
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src", "sub", "test.mp4")
	dst := filepath.Join(tmpDir, "dst", "test.mp4")

	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("hello movie content")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := MoveFile(src, dst); err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Verify src is gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file should not exist after MoveFile")
	}

	// Verify dst exists with same content
	readBack, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed reading dst: %v", err)
	}
	if string(readBack) != string(content) {
		t.Errorf("content mismatch, got %q, expected %q", string(readBack), string(content))
	}
}
