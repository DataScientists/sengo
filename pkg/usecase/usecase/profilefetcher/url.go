package profilefetcher

import (
	"fmt"
	"strings"
)

// usernameFromURL extracts the LinkedIn username slug from a profile URL.
//
//	"https://www.linkedin.com/in/some-username/" -> "some-username"
//
// The slug is the identifier stored as linkedin_urn and passed to RapidAPI as
// the ?username= query parameter. Returns an error when the URL does not point
// to a linkedin.com/in/ profile or the slug is empty.
func usernameFromURL(profileURL string) (string, error) {
	u := strings.TrimSpace(profileURL)
	if u == "" {
		return "", fmt.Errorf("linkedin url is empty")
	}

	// Drop any query string or fragment.
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}

	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "www.")

	const marker = "linkedin.com/in/"
	idx := strings.Index(strings.ToLower(u), marker)
	if idx < 0 {
		return "", fmt.Errorf("not a linkedin profile url: %s", profileURL)
	}

	slug := u[idx+len(marker):]
	slug = strings.TrimSuffix(slug, "/")
	// Guard against trailing path segments such as /in/<slug>/detail.
	if i := strings.Index(slug, "/"); i >= 0 {
		slug = slug[:i]
	}
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return "", fmt.Errorf("could not extract username from url: %s", profileURL)
	}

	return slug, nil
}
