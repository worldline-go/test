# test 🧪

Testing database queries we should initialize a test database and put some data in it and try to run our process to see if it is working as expected.

```sh
go get github.com/worldline-go/test
```

## PostgreSQL

You need to have a running PostgreSQL database to run the tests. To do that run it in the package level test main function.

Our test package has `Main` function it will run the given function and accept a defer function for cleanup.

```go
// "github.com/worldline-go/test"
// "github.com/worldline-go/test/container"
// "github.com/worldline-go/test/testdb"

var GlobalDB *sqlx.DB

func TestMain(m *testing.M) {
	test.Main(m, func(ctx context.Context) (func(), error) {
		postgres, err := container.Postgres(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to start postgres container: %w", err)
		}

		db, err := sqlx.Connect("pgx", postgres.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}

		GlobalDB = db

		return func() {
			postgres.Close()
		}, nil
	})
}
```

After that you can use the `GlobalDB` variable in your tests.

`testdb` has some helper functions to create schema, execute files/folders, and drop schema.

Use like that in your test function.

```go
// "github.com/worldline-go/test/testdb"

func TestDatabase(t *testing.T) {
	tDB := testdb.NewDatabase(t, GlobalDB.DB)

	schemaName := tDB.NameGen("database")
	if err := tDB.CreateSchema(schemaName); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	defer tDB.DropSchema(schemaName)

	err := tDB.ExecuteFolder("../../migrations/awesome")
	require.NoError(t, err)

	t.Run("Get Record", func(t *testing.T) {
		tDB.ExecuteFiles([]string{"testdata/records.sql"})

		db := newDB(GlobalDB)

		record, err := db.GetRecord(context.Background(), 1234, 1)
		require.NoError(t, err)
		require.NotNil(t, record)

		// check data
		require.Equal(t, int64(1234), record.FileID)
		require.Equal(t, record.RecordDetails, types.RawJSON([]byte(`{"file_id":1234,"record_id":1}`)))
	})
}
```
