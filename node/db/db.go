package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_log "github.com/cryptopunkscc/astrald/log"
	"modernc.org/ql"
	"sync"
)

type Database struct {
	mu sync.Mutex
	*sql.DB
}

var log = _log.Tag("db")

func NewFileDatabase(dbFile string) (*Database, error) {
	db, err := sql.Open("ql2", dbFile)
	if err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func NewMemDatabase(name string) (*Database, error) {
	db, err := sql.Open("ql-mem", name)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

// TxDoContext begins a database transaction and passes the tx to the provided function f. If f returns with an error,
// database transaction is rolled back, otherwise it gets committed. TxDoContext returns the error returned by f.
func (db *Database) TxDoContext(ctx context.Context, f func(tx *sql.Tx) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// TxDo invokes TxDoContext with background context.
func (db *Database) TxDo(f func(tx *sql.Tx) error) error {
	return db.TxDoContext(context.Background(), f)
}

// GetSchema returns the schema of the provided table.
func (db *Database) GetSchema(tableName string) (schema string, err error) {
	db.TxDo(func(tx *sql.Tx) error {
		schema, err = db.getSchema(tx, tableName)
		return nil
	})
	return
}

func (db *Database) getSchema(tx *sql.Tx, tableName string) (schema string, err error) {
	var name string
	rows, err := tx.Query("SELECT * FROM __Table WHERE Name == $1", tableName)
	if err != nil {
		return "", err
	}

	if !rows.Next() {
		return "", errors.New("not found")
	}

	err = rows.Scan(&name, &schema)
	rows.Close()
	return
}

func (db *Database) CreateTable(tableName string, fields string) error {
	return db.TxDo(func(tx *sql.Tx) error {
		query := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, fields)

		log.Logv(1, "%s", query)

		_, err := tx.Exec(query)
		return err
	})
}

func init() {
	ql.RegisterDriver2()
}
