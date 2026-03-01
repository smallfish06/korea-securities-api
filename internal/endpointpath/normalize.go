package endpointpath

import "strings"

// Normalize ensures a non-empty path has a leading slash and broker prefix.
func Normalize(path, prefix, prefixSlash string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, prefixSlash) {
		path = prefix + path
	}
	return path
}
