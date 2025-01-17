package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"testing"
)

type Database struct {
	DB *sql.DB
	T  *testing.T

	// schema counter
	schemaCounter int32
}

func NewDatabase(t *testing.T, db *sql.DB) *Database {
	return &Database{
		DB: db,
		T:  t,
	}
}

func (db *Database) NameGen(prefix string) string {
	atomic.AddInt32(&db.schemaCounter, 1)

	return prefix + "_" + strconv.Itoa(int(db.schemaCounter))
}

func (db *Database) SetSchema(schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	db.T.Helper()
	db.T.Logf("set schema to %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "SET search_path TO "+schema)
	return err
}

func (db *Database) CreateSchema(schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	db.T.Helper()
	db.T.Logf("create schema %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "CREATE SCHEMA "+schema)
	return err
}

func (db *Database) DropSchema(schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	db.T.Helper()
	db.T.Logf("drop schema %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	return err
}

func (db *Database) ExecuteFolder(folder string, opts ...OptionExec) error {
	dirEntry, err := os.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("could not read folder %s: %w", folder, err)
	}

	var files []string
	for _, file := range dirEntry {
		if file.IsDir() {
			continue
		}

		files = append(files, path.Join(folder, file.Name()))
	}

	return db.ExecuteFiles(files, opts...)
}

func (db *Database) ExecuteFiles(files []string, opts ...OptionExec) error {
	opt := apply(opts)
	db.T.Helper()

	for _, file := range files {
		db.T.Logf("execute file %s", file)

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("could not read file %s: %w", file, err)
		}

		contentStr := string(content)
		if len(opt.Values) > 0 {
			contentStr = os.Expand(contentStr, func(key string) string {
				return opt.Values[key]
			})
		}

		ctx := opt.Ctx
		if opt.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
			defer cancel()
		}

		if _, err = db.DB.ExecContext(ctx, contentStr); err != nil {
			return err
		}
	}

	return nil
}
