package testdb

import (
	"context"
	"database/sql"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"testing"
)

type Database struct {
	DB *sql.DB

	// schema counter
	schemaCounter int32
}

func NewDatabase(t *testing.T, db *sql.DB) *Database {
	return &Database{
		DB: db,
	}
}

func (db *Database) NameGen(prefix string) string {
	atomic.AddInt32(&db.schemaCounter, 1)

	return prefix + "_" + strconv.Itoa(int(db.schemaCounter))
}

func (db *Database) SetSchema(t *testing.T, schema string, opts ...OptionContext) {
	opt := apply(opts)
	schema = trim(schema)

	t.Helper()
	t.Logf("set schema to %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "SET search_path TO "+schema)
	if err != nil {
		t.Fatalf("could not set schema to %s: %v", schema, err)
	}
}

func (db *Database) CreateSchema(t *testing.T, schema string, opts ...OptionContext) {
	opt := apply(opts)
	schema = trim(schema)

	t.Helper()
	t.Logf("create schema %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "CREATE SCHEMA "+schema)
	if err != nil {
		t.Fatalf("could not create schema %s: %v", schema, err)
	}
}

func (db *Database) DropSchema(t *testing.T, schema string, opts ...OptionContext) {
	opt := apply(opts)
	schema = trim(schema)

	t.Helper()
	t.Logf("drop schema %s", schema)

	_, err := db.DB.ExecContext(opt.Ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	if err != nil {
		t.Fatalf("could not drop schema %s: %v", schema, err)
	}
}

func (db *Database) ExecuteFolder(t *testing.T, folder string, opts ...OptionExec) {
	dirEntry, err := os.ReadDir(folder)
	if err != nil {
		t.Fatalf("could not read folder %s: %v", folder, err)
	}

	var files []string
	for _, file := range dirEntry {
		if file.IsDir() {
			continue
		}

		files = append(files, path.Join(folder, file.Name()))
	}

	db.ExecuteFiles(t, files, opts...)
}

func (db *Database) ExecuteFiles(t *testing.T, files []string, opts ...OptionExec) {
	opt := apply(opts)
	t.Helper()

	for _, file := range files {
		t.Logf("execute file %s", file)

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("could not read file %s: %v", file, err)
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
			t.Fatalf("could not execute file %s: %v", file, err)
		}
	}
}
