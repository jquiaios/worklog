package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/jquiaios/worklog/internal/entry"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS entries (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	type       TEXT    NOT NULL CHECK(type IN ('highlight', 'lowlight')),
	body       TEXT    NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

type DB struct {
	sql *sql.DB
}

func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return &DB{sql: conn}, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) Insert(e entry.Entry) (int64, error) {
	res, err := d.sql.Exec(
		`INSERT INTO entries (type, body, created_at) VALUES (?, ?, ?)`,
		string(e.Type),
		e.Body,
		e.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) List(typeFilter string) ([]entry.Entry, error) {
	q := `SELECT id, type, body, created_at FROM entries`
	var args []any
	if typeFilter != "" {
		q += ` WHERE type = ?`
		args = append(args, typeFilter)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := d.sql.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []entry.Entry
	for rows.Next() {
		var e entry.Entry
		var t, createdAt string
		if err := rows.Scan(&e.ID, &t, &e.Body, &createdAt); err != nil {
			return nil, err
		}
		e.Type = entry.Type(t)
		e.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			e.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				return nil, err
			}
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (d *DB) Update(id int64, body string) error {
	_, err := d.sql.Exec(`UPDATE entries SET body = ? WHERE id = ?`, body, id)
	return err
}

func (d *DB) Delete(id int64) (bool, error) {
	res, err := d.sql.Exec(`DELETE FROM entries WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
