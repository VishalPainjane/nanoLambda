package registry

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// function represents a deployed serverless function
type Function struct {
	Name        string
	Runtime     string
	ImageTag    string
	CreatedAt   time.Time
	MemoryLimit int64
	Timeout     int
}

// manager handles database interactions
type Manager struct {
	db *sql.DB
}

// newmanager initializes the registry with a sqlite database
func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	m := &Manager{db: db}
	if err := m.initSchema(); err != nil {
		return nil, err
	}

	return m, nil
}

// initschema creates the necessary tables
func (m *Manager) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS functions (
		name TEXT PRIMARY KEY,
		runtime TEXT,
		image_tag TEXT,
		created_at DATETIME,
		memory_limit INTEGER,
		timeout INTEGER
	);`
	_, err := m.db.Exec(query)
	return err
}

// registerfunction adds or updates a function in the registry
func (m *Manager) RegisterFunction(fn Function) error {
	query := `
	INSERT INTO functions (name, runtime, image_tag, created_at, memory_limit, timeout)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(name) DO UPDATE SET
		runtime=excluded.runtime,
		image_tag=excluded.image_tag,
		memory_limit=excluded.memory_limit,
		timeout=excluded.timeout;
	`
	_, err := m.db.Exec(query, fn.Name, fn.Runtime, fn.ImageTag, fn.CreatedAt, fn.MemoryLimit, fn.Timeout)
	return err
}

// getfunction retrieves a function by name
func (m *Manager) GetFunction(name string) (*Function, error) {
	query := `SELECT name, runtime, image_tag, created_at, memory_limit, timeout FROM functions WHERE name = ?`
	row := m.db.QueryRow(query, name)

	var fn Function
	err := row.Scan(&fn.Name, &fn.Runtime, &fn.ImageTag, &fn.CreatedAt, &fn.MemoryLimit, &fn.Timeout)
	if err != nil {
		return nil, err
	}
	return &fn, nil
}

// listfunctions returns all registered functions
func (m *Manager) ListFunctions() ([]Function, error) {
	query := `SELECT name, runtime, image_tag, created_at, memory_limit, timeout FROM functions`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var functions []Function
	for rows.Next() {
		var fn Function
		if err := rows.Scan(&fn.Name, &fn.Runtime, &fn.ImageTag, &fn.CreatedAt, &fn.MemoryLimit, &fn.Timeout); err != nil {
			return nil, err
		}
		functions = append(functions, fn)
	}
	return functions, nil
}

// close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}