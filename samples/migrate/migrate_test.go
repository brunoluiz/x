package migrate_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/brunoluiz/x/database/migrate/v2"
	"github.com/brunoluiz/x/database/sqlite/v2"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
)

func TestRunKeepsSQLiteDatabaseOpen(t *testing.T) {
	db, err := sqlite.New(context.Background(), t.TempDir()+"/migrate.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Get().Close() })

	fs := fstest.MapFS{
		"000001_create_users.up.sql": &fstest.MapFile{Data: []byte(`create table users (
		  id integer primary key,
		  name text not null
		)`)},
	}

	driver, err := sqlite3.WithInstance(db.Get(), &sqlite3.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if runErr := migrate.Run(driver, fs); runErr != nil {
		t.Fatalf("Run() error = %v", runErr)
	}

	if pingErr := db.Get().PingContext(context.Background()); pingErr != nil {
		t.Fatalf("PingContext() error = %v", pingErr)
	}

	if _, execErr := db.Get().ExecContext(context.Background(), `insert into users (name) values (?)`, "alice"); execErr != nil {
		t.Fatalf("ExecContext() error = %v", execErr)
	}
}
