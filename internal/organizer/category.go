package organizer

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Matches leading tags like [1080p], [YTS], [RARBG], {tags}
	leadingBracketTagRegex = regexp.MustCompile(`^(\[[^\]]*\]|\{[^\}]*\})\s*`)
	// Matches leading year in parentheses like (1999) or (2001) when followed by text
	leadingYearTagRegex = regexp.MustCompile(`^\((?:19|20)\d{2}\)\s+`)
)

// StandardCategories returns the complete list of 27 standard categories:
// "#" for numbers and symbols, followed by "a" through "z".
func StandardCategories() []string {
	cats := make([]string, 0, 27)
	cats = append(cats, "#")
	for c := 'a'; c <= 'z'; c++ {
		cats = append(cats, string(c))
	}
	return cats
}

// IsCategoryDir returns true if the folder name corresponds to one of the 27 categories
// ("#" or any single letter a-z, case-insensitive).
func IsCategoryDir(name string) bool {
	if name == "#" {
		return true
	}
	if len(name) == 1 {
		b := strings.ToLower(name)[0]
		return b >= 'a' && b <= 'z'
	}
	return false
}

// CategoryOptions configures how movie categories are determined.
type CategoryOptions struct {
	IgnoreArticles  bool
	IgnoredArticles []string
}

// GetCategory returns the category name ("#" or "a"-"z") for a given movie filename,
// using default categorization rules without article stripping.
func GetCategory(filename string) string {
	return GetCategoryWithOptions(filename, CategoryOptions{})
}

// GetCategoryWithOptions returns the category name ("#" or "a"-"z") for a given movie filename.
// It strips file extension, leading tags like [1080p] or (2001), brackets, spaces,
// optionally strips leading articles (e.g. "The", "El"), and normalizes accents.
func GetCategoryWithOptions(filename string, opts CategoryOptions) string {
	// Strip directory and extension first so .mp4 does not count as title characters
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	clean := strings.TrimSuffix(base, ext)
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return "#"
	}

	// Remove leading bracket tags like [1080p] or [YTS] if followed by a title
	for {
		stripped := leadingBracketTagRegex.ReplaceAllString(clean, "")
		if stripped != clean && strings.TrimSpace(stripped) != "" {
			clean = strings.TrimSpace(stripped)
		} else {
			break
		}
	}

	// Remove leading year in parentheses if followed by title, e.g. "(2001) A Space Odyssey"
	if leadingYearTagRegex.MatchString(clean) {
		clean = strings.TrimSpace(leadingYearTagRegex.ReplaceAllString(clean, ""))
	}

	// Strip leading articles if configured (e.g. "The Matrix" -> "Matrix", "El Padrino" -> "Padrino")
	if opts.IgnoreArticles && len(opts.IgnoredArticles) > 0 {
		clean = stripLeadingArticle(clean, opts.IgnoredArticles)
	}

	// Skip leading punctuation, brackets, quotes, etc.
	runes := []rune(clean)
	var firstRune rune
	found := false

	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			firstRune = r
			found = true
			break
		}
	}

	if !found {
		return "#"
	}

	if unicode.IsDigit(firstRune) {
		return "#"
	}

	// Normalize diacritics / accents (e.g. Á -> a, Ñ -> n)
	normalized := normalizeRune(firstRune)
	if normalized >= 'a' && normalized <= 'z' {
		return string(normalized)
	}

	return "#"
}

// stripLeadingArticle checks if clean starts with any configured article followed by a delimiter
// (space, dot, underscore, dash) and removes it.
func stripLeadingArticle(title string, articles []string) string {
	titleLower := strings.ToLower(title)
	for _, art := range articles {
		art = strings.TrimSpace(art)
		if art == "" {
			continue
		}
		artLower := strings.ToLower(art)
		if strings.HasPrefix(titleLower, artLower) {
			rest := title[len(art):]
			if len(rest) > 0 {
				firstChar := rest[0]
				if firstChar == ' ' || firstChar == '.' || firstChar == '_' || firstChar == '-' {
					stripped := strings.TrimSpace(strings.TrimPrefix(rest, string(firstChar)))
					if stripped != "" {
						return stripped
					}
				}
			}
		}
	}
	return title
}

// normalizeRune converts uppercase and accented letters to lowercase ASCII a-z.
func normalizeRune(r rune) rune {
	r = unicode.ToLower(r)

	switch r {
	case 'á', 'à', 'â', 'ä', 'ã', 'å', 'æ':
		return 'a'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'ó', 'ò', 'ô', 'ö', 'õ', 'ø':
		return 'o'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'ñ':
		return 'n'
	case 'ç':
		return 'c'
	default:
		return r
	}
}
