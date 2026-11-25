package main

import (
	"os"
	"path/filepath"
)

const (
	defaultDataDir   = "_data"
	defaultPortalURL = "http://localhost:9090"
)

func dataDir() string {
	if dir := os.Getenv("RELEASEDESK_DATA_DIR"); dir != "" {
		return dir
	}
	return defaultDataDir
}

func databasePath() string {
	return filepath.Join(dataDir(), "releasedesk.db")
}

func portalPublicURL() string {
	if url := os.Getenv("RELEASEDESK_PORTAL_URL"); url != "" {
		return url
	}
	return defaultPortalURL
}
