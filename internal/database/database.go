package database

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err = ensureTransactionReversalColumn(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; " + schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return db, nil
}

func ensureTransactionReversalColumn(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operator_id INTEGER NOT NULL REFERENCES users(id),
		kind TEXT NOT NULL CHECK (kind IN ('income','expense')),
		amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
		category TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		occurred_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		reversal_of_id INTEGER REFERENCES transactions(id)
	)`); err != nil {
		return err
	}
	rows, err := db.Query(`PRAGMA table_info(transactions)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	hasColumn := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == "reversal_of_id" {
			hasColumn = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !hasColumn {
		_, err = db.Exec(`ALTER TABLE transactions ADD COLUMN reversal_of_id INTEGER REFERENCES transactions(id)`)
	}
	return err
}
