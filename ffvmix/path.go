package ffvmix

import (
	"net/url"
	"path/filepath"
)

func isAbsoluteLocalPath(path string) bool {
	return filepath.IsAbs(path)
}

func cleanLocalPath(path string) string {
	return filepath.Clean(path)
}

func joinLocalPath(base, path string) string {
	return filepath.Join(base, path)
}

func hasURIScheme(path string) bool {
	parsed, err := url.Parse(path)
	return err == nil && parsed.Scheme != ""
}
