package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	// Create manifests directory if it doesn't exist
	manifestsDir := "manifests"
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create manifests directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(manifestsDir, "shipyard.db")

	// Open database connection
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the database tables if they don't exist
func (db *DB) initSchema() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// GetPath returns the database file path
func (db *DB) GetPath() string {
	return db.path
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// GetOrCreateApp gets an app by name or creates it if it doesn't exist
func (db *DB) GetOrCreateApp(name string) (int64, error) {
	// First try to get existing app
	var appID int64
	err := db.conn.QueryRow("SELECT id FROM apps WHERE name = ?", name).Scan(&appID)
	if err == nil {
		return appID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query app: %w", err)
	}

	// App doesn't exist, create it
	result, err := db.conn.Exec("INSERT INTO apps (name) VALUES (?)", name)
	if err != nil {
		return 0, fmt.Errorf("failed to create app: %w", err)
	}

	appID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get app ID: %w", err)
	}

	return appID, nil
}

// BeginTx starts a new database transaction
func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.conn.Begin()
}