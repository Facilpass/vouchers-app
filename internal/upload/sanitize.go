package upload

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var accentStripper = transform.Chain(
	norm.NFD,
	transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}),
	norm.NFC,
)

// injectionChars are shell/SQL metacharacters that indicate intentional injection.
// When present, whitespace is also stripped (not converted to underscore).
var injectionChars = regexp.MustCompile(`[;'"\x60<>\\|&$]`)

var dangerousChars = regexp.MustCompile(`[/\\:;|&$` + "`" + `'"<>(){}\[\]*?!#%=+,~@^]`)
var nonWhitelisted = regexp.MustCompile(`[^A-Za-z0-9._-]`)
var multiUnderscore = regexp.MustCompile(`_{2,}`)
var multiHyphen = regexp.MustCompile(`-{3,}`)
var leadingDots = regexp.MustCompile(`^\.+`)
var doubleDots = regexp.MustCompile(`\.\.`)
var leadingUnderscores = regexp.MustCompile(`^_+`)

var pathRegex = regexp.MustCompile(`^[a-z0-9-]+(/[a-z0-9-]+)*$`)
var validExts = map[string]bool{
	".jpeg": true,
	".jpg":  true,
	".png":  true,
	".webp": true,
}

const (
	maxFilenameLen = 128
	maxPathDepth   = 10
)

// SanitizeFilename normalizes a filename, preserving original case.
// Accent characters are stripped to ASCII equivalents.
// Spaces become underscores unless the input contains injection metacharacters,
// in which case spaces are removed entirely along with the dangerous chars.
func SanitizeFilename(in string) string {
	s, _, _ := transform.String(accentStripper, in)

	// Remove control characters (0x00–0x1F, 0x7F).
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, s)

	hasInjection := injectionChars.MatchString(s)

	s = dangerousChars.ReplaceAllString(s, "")

	if hasInjection {
		// Strip whitespace — it was surrounding injection chars.
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\t", "")
	} else {
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "\t", "_")
	}

	s = nonWhitelisted.ReplaceAllString(s, "")
	s = multiUnderscore.ReplaceAllString(s, "_")
	s = multiHyphen.ReplaceAllString(s, "-")
	s = doubleDots.ReplaceAllString(s, "")
	s = leadingDots.ReplaceAllString(s, "")
	s = leadingUnderscores.ReplaceAllString(s, "")

	if len(s) > maxFilenameLen {
		s = s[:maxFilenameLen]
	}
	return s
}

// SanitizePath normalizes a storage path to lowercase slug segments separated by "/".
// Each segment may contain only [a-z0-9-]. Leading/trailing slashes and empty
// segments (from path traversal sequences like "../..") are removed.
func SanitizePath(in string) string {
	s, _, _ := transform.String(accentStripper, in)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "\t", "-")

	parts := strings.Split(s, "/")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, p)
		p = multiHyphen.ReplaceAllString(p, "-")
		p = strings.Trim(p, "-")
		if p != "" {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, "/")
}

// IsValidPath reports whether p is a sanitized path: 1–10 lowercase segments
// of [a-z0-9-], joined by single slashes, no leading/trailing slash.
func IsValidPath(p string) bool {
	if p == "" {
		return false
	}
	if !pathRegex.MatchString(p) {
		return false
	}
	depth := strings.Count(p, "/") + 1
	return depth <= maxPathDepth
}

// IsValidExt reports whether ext is an allowed image extension (.jpeg, .jpg, .png, .webp).
func IsValidExt(ext string) bool {
	return validExts[strings.ToLower(ext)]
}
