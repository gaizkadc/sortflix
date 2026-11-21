package organizer

import "testing"

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"lowercase simple", "american ninja.mp4", "a"},
		{"uppercase simple", "Batman Begins.mkv", "b"},
		{"number prefix", "1917 (2019).mkv", "#"},
		{"number in word", "300.mp4", "#"},
		{"square brackets prefix", "[1080p] The Matrix.mkv", "t"},
		{"parentheses prefix", "(2001) A Space Odyssey.mp4", "a"},
		{"dots and spaces prefix", "...Casablanca.avi", "c"},
		{"underscore prefix", "_Dune.mkv", "d"},
		{"dash prefix", "- Fargo.mp4", "f"},
		{"accented Á", "Ágora (2009).mkv", "a"},
		{"accented É", "Élite.mp4", "e"},
		{"accented Í", "Ícaro.mkv", "i"},
		{"accented Ó", "Ópera.avi", "o"},
		{"accented Ú", "Último tren a Busan.mkv", "u"},
		{"letter Ñ", "Ñu.mkv", "n"},
		{"only symbols", "[--]__....mp4", "#"},
		{"empty string", "", "#"},
		{"number 8", "8 Millas.mp4", "#"},
		{"special character", "¡Ay, Carmela!.mkv", "a"},
		{"german umlaut", "Überlingen.mp4", "u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategory(tt.filename)
			if got != tt.expected {
				t.Errorf("GetCategory(%q) = %q, expected %q", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestIsCategoryDir(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"#", true},
		{"a", true},
		{"z", true},
		{"A", true},
		{"Z", true},
		{"aa", false},
		{"american ninja", false},
		{"12", false},
		{"", false},
		{"#1", false},
	}

	for _, tt := range tests {
		got := IsCategoryDir(tt.name)
		if got != tt.expected {
			t.Errorf("IsCategoryDir(%q) = %v, expected %v", tt.name, got, tt.expected)
		}
	}
}
