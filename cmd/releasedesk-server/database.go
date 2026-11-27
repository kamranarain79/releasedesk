package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brewbits-co/releasedesk/pkg/storage"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

// Initialize SQLite by default, or PostgreSQL when RELEASEDESK_DATABASE_URL is set.
func newDBEngine() (*xorm.Engine, error) {
	storage.SetRoot(dataDir())

	var (
		engine *xorm.Engine
		err    error
	)
	if url := os.Getenv("RELEASEDESK_DATABASE_URL"); url != "" {
		engine, err = xorm.NewEngine("postgres", url)
	} else {
		if err = os.MkdirAll(dataDir(), 0o755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)", databasePath())
		engine, err = xorm.NewEngine("sqlite3", dsn)
	}
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

	if engine.DriverName() == "sqlite3" {
		if _, err = engine.Exec("PRAGMA journal_mode = WAL;"); err != nil {
			log.Fatalf("Failed to enable WAL mode: %v", err)
		}
	}

	return engine, nil
}
