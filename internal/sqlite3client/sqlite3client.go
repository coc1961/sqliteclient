package sqlite3client

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite3Client interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryCallback(fn SqlFN, q string, args ...interface{}) error
	Close() error
}

type SqlFN func(columns []string, row []interface{}) error

type sqlite3Client struct {
	db *sqlx.DB
}

func New(connStr string) (Sqlite3Client, error) {
	driver := "sqlite3"

	db, err := sqlx.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	return &sqlite3Client{db: db}, nil
}

func (s *sqlite3Client) Close() error {
	return s.db.Close()
}
func (s *sqlite3Client) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *sqlite3Client) QueryCallback(fn SqlFN, q string, args ...interface{}) error {
	if fn == nil {
		_, err := s.Exec(q, args...)
		return err
	}

	rows, err := s.db.Queryx(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	for rows.Rows.Next() {
		r, err := rows.SliceScan()
		if err != nil {
			return err
		}
		err = fn(columns, r)
		if err != nil {
			return err
		}
	}
	return nil
}
