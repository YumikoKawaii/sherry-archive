package slug

import (
	"fmt"

	goslug "github.com/gosimple/slug"
)

// Generate creates a URL-friendly slug from a title, appending a suffix for uniqueness.
func Generate(title, suffix string) string {
	base := goslug.Make(title)
	if suffix == "" {
		return base
	}
	return fmt.Sprintf("%s-%s", base, suffix)
}

// Make returns a plain slug without suffix.
func Make(title string) string {
	return goslug.Make(title)
}
