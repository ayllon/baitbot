package markov

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type (
	Parameters struct {
		PrefixLen int
		DryRun    bool
	}

	Markov struct {
		params  Parameters
		db      *sql.DB
		dbMutex sync.Mutex
	}
)

// New returns a new Markov instance
// path is a sqlite file (or :memory:)
func New(path string, params Parameters) (*Markov, error) {
	log.Debug("Opening SQLite ", path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	m := &Markov{
		params: params,
		db:     db,
	}
	return m, m.createSchema()
}

// Close closes the underlying database
func (m *Markov) Close() error {
	return m.db.Close()
}

// createSchema creates the required tables if not exist
func (m *Markov) createSchema() error {
	_, err := m.db.Exec(`
CREATE TABLE IF NOT EXISTS markov_prefix (prefix TEXT)`)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS markov_prefix_index ON markov_prefix(prefix)`)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(`
CREATE TABLE IF NOT EXISTS markov_next
	(prefix_rowid INTEGER NOT NULL, token TEXT,
	 FOREIGN KEY (prefix_rowid) REFERENCES markov_prefix(rowid))`)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(`CREATE TABLE IF NOT EXISTS markov_start (prefix TEXT)`)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS markov_start_index ON markov_start(prefix)`)
	return err
}

// Clear removes all stored information. Beware!
func (m *Markov) Clear() error {
	if _, err := m.db.Exec("DELETE FROM markov_start"); err != nil {
		return err
	}
	if _, err := m.db.Exec("DELETE FROM markov_next"); err != nil {
		return err
	}
	if _, err := m.db.Exec("DELETE FROM markov_prefix"); err != nil {
		return err
	}
	return nil
}
