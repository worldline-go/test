package testdb

import "database/sql"

type Database struct {
	DB sql.DB
}

type Option struct{}
