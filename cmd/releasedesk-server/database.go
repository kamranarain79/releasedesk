package main

import (
	"log"
	"os"

	"github.com/brewbits-co/releasedesk/pkg/storage"
	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

// newDBEngine connects to PostgreSQL and applies migrations.
// Artifacts stay on disk under the data directory.
func newDBEngine() (*xorm.Engine, error) {
	storage.SetRoot(dataDir())
	if err := os.MkdirAll(dataDir(), 0o755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	engine, err := xorm.NewEngine("postgres", databaseURL())
	if err != nil {
		log.Fatalf("Failed to start database engine: %v", err)
	}

	if err = engine.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	engine.SetMapper(names.GonicMapper{})
	engine.ShowSQL(true)

	if err = applyMigrations(engine); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	return engine, nil
}
