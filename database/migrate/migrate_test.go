package migrate_test

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	"github.com/brunoluiz/x/database/migrate"
	_ "modernc.org/sqlite"
)

type testDB struct {
	db       *sql.DB
	typeName string
}

func (d testDB) Get() *sql.DB { return d.db }

func (d testDB) Type() string { return d.typeName }

func TestRunKeepsSQLiteDatabaseOpen(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/migrate.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	fs := fstest.MapFS{
		"000001_create_users.up.sql": &fstest.MapFile{Data: []byte(`create table users (
		  id integer primary key,
		  name text not null
		)`)},
	}

	if runErr := migrate.Run(testDB{db: db, typeName: "sqlite"}, fs); runErr != nil {
		t.Fatalf("Run() error = %v", runErr)
	}

	if pingErr := db.PingContext(context.Background()); pingErr != nil {
		t.Fatalf("PingContext() error = %v", pingErr)
	}

	if _, execErr := db.ExecContext(context.Background(), `insert into users (name) values (?)`, "alice"); execErr != nil {
		t.Fatalf("ExecContext() error = %v", execErr)
	}
}
