package util

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const (
	PublicDir = "public"
	PostsDir  = "posts"
	SourceDir = "posts"
	StaticDir = "static"
)

func Slugify(s string) string {
	// Normalize the string to decompose combined characters
	s = norm.NFKD.String(s)

	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove all non-alphanumeric characters except hyphens and spaces
	reg := regexp.MustCompile("[^a-z0-9-\\s]")
	s = reg.ReplaceAllString(s, "")

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove any leading or trailing hyphens
	s = strings.Trim(s, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	// Limit the slug length to 100 characters, breaking at word boundaries if possible
	if len(s) > 100 {
		words := strings.Split(s, "-")
		s = ""
		for _, word := range words {
			if len(s)+len(word)+1 > 100 {
				break
			}
			if s != "" {
				s += "-"
			}
			s += word
		}
	}

	return s
}
