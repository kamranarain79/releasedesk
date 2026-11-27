package main

import (
	"os"
)

const (
	defaultDataDir   = "_data"
	defaultPortalURL = "http://localhost:9090"
	// Local Compose publishes Postgres on this URL. Docker overrides it with the service hostname.
	defaultDatabaseURL = "postgres://releasedesk:releasedesk@127.0.0.1:5432/releasedesk?sslmode=disable"
)

func dataDir() string {
	if dir := os.Getenv("RELEASEDESK_DATA_DIR"); dir != "" {
		return dir
	}
	return defaultDataDir
}

func databaseURL() string {
	if url := os.Getenv("RELEASEDESK_DATABASE_URL"); url != "" {
		return url
	}
	return defaultDatabaseURL
}

func portalPublicURL() string {
	if url := os.Getenv("RELEASEDESK_PORTAL_URL"); url != "" {
		return url
	}
	return defaultPortalURL
}
