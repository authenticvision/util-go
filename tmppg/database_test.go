package tmppg

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const ERRCODE_INVALID_CATALOG_NAME = "3D000"

func TestInstance_WithDatabase_Cleanup(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	a := assert.New(t)
	err := WithPostgresql(func(socketDir string) error {
		pg := NewInstance("host=" + socketDir)
		var dbname string
		err := pg.WithDatabase(t.Context(), func(pool *pgxpool.Pool) error {
			row := pool.QueryRow(t.Context(), "SELECT current_database();")
			err := row.Scan(&dbname)
			r.NoError(err)
			return errors.New("test error")
		})
		a.Error(err)
		_, err = pgx.Connect(t.Context(), "host="+socketDir+" dbname="+dbname)
		var pgError *pgconn.PgError
		a.ErrorAs(err, &pgError)
		a.Equal(ERRCODE_INVALID_CATALOG_NAME, pgError.Code)
		return nil
	})
	r.NoError(err)
}

func TestInstance_WithDatabase_Panic(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	a := assert.New(t)
	err := WithPostgresql(func(socketDir string) error {
		pg := NewInstance("host=" + socketDir)
		var dbname string
		err := pg.WithDatabase(t.Context(), func(pool *pgxpool.Pool) error {
			panic("test panic")
		})
		a.Error(err)
		_, err = pgx.Connect(t.Context(), "host="+socketDir+" dbname="+dbname)
		var pgError *pgconn.PgError
		a.ErrorAs(err, &pgError)
		a.Equal(ERRCODE_INVALID_CATALOG_NAME, pgError.Code)
		return nil
	})
	r.NoError(err)
}

func TestInstance_WithDatabase_RuntimeGoexit(t *testing.T) {
	t.Skip("this test will always fail. execute manually to verify cleanup behavior.")
	t.Parallel()
	_ = WithPostgresql(func(socketDir string) error {
		pg := NewInstance("host=" + socketDir)
		_ = pg.WithDatabase(t.Context(), func(pool *pgxpool.Pool) error {
			println("exiting now via t.FailNow()")
			t.FailNow()
			return nil
		})
		return nil
	})
}

const schemaSQL = `
CREATE TABLE test (
	id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	name TEXT NOT NULL
);

INSERT INTO test (name) VALUES ('test');
`

func TestInstance_WithDatabaseSchema(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	a := assert.New(t)
	err := WithPostgresql(func(socketDir string) error {
		pg := NewInstance("host=" + socketDir)
		err := pg.WithDatabaseSchema(t.Context(), schemaSQL, func(pool *pgxpool.Pool) error {
			var id int64
			var name, dbname string
			row := pool.QueryRow(t.Context(), "SELECT id, name, current_database() FROM test;")
			err := row.Scan(&id, &name, &dbname)
			if err != nil {
				return err
			}
			a.Equal(int64(1), id)
			a.Equal("test", name)
			a.Contains(dbname, "test")
			return err
		})
		return err
	})
	r.NoError(err)
}
