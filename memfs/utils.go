package memfs

import "strings"

// SplitPath splits path in segments
func SplitPath(path string) []string {
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, "/")
	if path == "" || path == "." {
		return []string{path}
	}

	if len(path) > 0 && !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "./") {
		path = "./" + path
	}

	return strings.Split(path, "/")
}
