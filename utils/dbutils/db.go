package dbutils

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

	// schema counter
	schemaCounter int32
}

type DatabaseTest struct {
	db *Database
}

func New(db *sql.DB) *Database {
	return &Database{
		DB: db,
	}
}

func NewTest(t *testing.T, db *sql.DB) *DatabaseTest {
	t.Helper()

	return &DatabaseTest{
		db: New(db),
	}
}

func (db *DatabaseTest) NameGen(prefix string) string {
	return db.db.NameGen(prefix)
}

func (db *Database) NameGen(prefix string) string {
	atomic.AddInt32(&db.schemaCounter, 1)

	return prefix + "_" + strconv.Itoa(int(db.schemaCounter))
}

func (db *DatabaseTest) SetSchema(t *testing.T, schema string, opts ...OptionContext) {
	t.Helper()

	if err := db.db.setSchema(t, schema, opts...); err != nil {
		t.Fatal(err)
	}
}

func (db *Database) SetSchema(schema string, opts ...OptionContext) error {
	return db.setSchema(nil, schema, opts...)
}

func (db *Database) setSchema(t *testing.T, schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	if t != nil {
		t.Helper()
		t.Logf("set schema to %s", schema)
	}

	_, err := db.DB.ExecContext(opt.Ctx, "SET search_path TO "+schema)
	if err != nil {
		return fmt.Errorf("could not set schema to %s: %w", schema, err)
	}

	return nil
}

func (db *DatabaseTest) CreateSchema(t *testing.T, schema string, opts ...OptionContext) {
	t.Helper()

	if err := db.db.createSchema(t, schema, opts...); err != nil {
		t.Fatal(err)
	}
}

func (db *Database) CreateSchema(schema string, opts ...OptionContext) error {
	return db.createSchema(nil, schema, opts...)
}

func (db *Database) createSchema(t *testing.T, schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	if t != nil {
		t.Helper()
		t.Logf("create schema %s", schema)
	}

	_, err := db.DB.ExecContext(opt.Ctx, "CREATE SCHEMA "+schema)
	if err != nil {
		return fmt.Errorf("could not create schema %s: %w", schema, err)
	}

	return nil
}

func (db *DatabaseTest) DropSchema(t *testing.T, schema string, opts ...OptionContext) {
	t.Helper()

	if err := db.db.dropSchema(t, schema, opts...); err != nil {
		t.Fatal(err)
	}
}

func (db *Database) DropSchema(schema string, opts ...OptionContext) error {
	return db.dropSchema(nil, schema, opts...)
}

func (db *Database) dropSchema(t *testing.T, schema string, opts ...OptionContext) error {
	opt := apply(opts)
	schema = trim(schema)

	if t != nil {
		t.Helper()
		t.Logf("drop schema %s", schema)
	}

	_, err := db.DB.ExecContext(opt.Ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	if err != nil {
		return fmt.Errorf("could not drop schema %s: %w", schema, err)
	}

	return nil
}

func (db *DatabaseTest) ExecuteFolder(t *testing.T, folder string, opts ...OptionExec) {
	t.Helper()

	if err := db.db.executeFolder(t, folder, opts...); err != nil {
		t.Fatal(err)
	}
}

func (db *Database) ExecuteFolder(folder string, opts ...OptionExec) error {
	return db.executeFolder(nil, folder, opts...)
}

func (db *Database) executeFolder(t *testing.T, folder string, opts ...OptionExec) error {
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

	return db.executeFiles(t, files, opts...)
}

func (db *DatabaseTest) ExecuteFiles(t *testing.T, files []string, opts ...OptionExec) {
	t.Helper()

	if err := db.db.executeFiles(t, files, opts...); err != nil {
		t.Fatal(err)
	}
}

func (db *Database) ExecuteFiles(files []string, opts ...OptionExec) error {
	return db.executeFiles(nil, files, opts...)
}

func (db *Database) executeFiles(t *testing.T, files []string, opts ...OptionExec) error {
	opt := apply(opts)
	if t != nil {
		t.Helper()
	}

	for _, file := range files {
		if t != nil {
			t.Logf("execute file %s", file)
		}

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
			return fmt.Errorf("could not execute file %s: %w", file, err)
		}
	}

	return nil
}
