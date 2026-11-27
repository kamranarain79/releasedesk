package storage

import (
	"path/filepath"
	"regexp"
)

// Root is the artifact directory. SetRoot points it at RELEASEDESK_DATA_DIR/storage.
var Root = filepath.Join("_data", "storage")

func SetRoot(dataDir string) {
	if dataDir == "" {
		dataDir = "_data"
	}
	Root = filepath.Join(dataDir, "storage")
}

var checksumPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

// ValidChecksum reports whether value is a SHA-256 hex digest safe to use as a path segment.
func ValidChecksum(value string) bool {
	return checksumPattern.MatchString(value)
}

// ConvertChecksumToPath generates a storage path for a file based on its checksum.
func ConvertChecksumToPath(checksum string) string {
	return filepath.Join(Root, checksum[:2], checksum)
}
