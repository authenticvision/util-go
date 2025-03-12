package tmppg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log/slog"
	"math/rand/v2"
)

type Instance struct {
	connString string
	log        *slog.Logger
}

func NewInstance(connString string) *Instance {
	return &Instance{
		connString: connString,
		log:        slog.Default(),
	}
}

func (i *Instance) WithDatabase(ctx context.Context, fn func(pool *pgxpool.Pool) error) (err error) {
	var conn *pgx.Conn
	conn, err = pgx.Connect(ctx, i.connString+" dbname=postgres")
	if err != nil {
		return fmt.Errorf("connect to admin database: %w", err)
	}
	defer conn.Close(ctx)

	dbname := fmt.Sprintf("test%d", rand.Uint32())
	i.log.Info("creating database", slog.String("name", dbname))
	_, err = conn.Exec(ctx, "CREATE DATABASE "+dbname)
	if err != nil {
		return fmt.Errorf("create database %q: %w", dbname, err)
	}

	// run database removal deferred, so the database also gets removed on
	// runtime.Goexit() and t.FailNow()
	defer func() {
		if recoveredError := recover(); recoveredError != nil {
			err = fmt.Errorf("panic: %v", recoveredError)
		}

		i.log.Info("dropping database", slog.String("name", dbname))
		_, dropError := conn.Exec(ctx, "DROP DATABASE "+dbname+" WITH (FORCE)")
		if dropError != nil {
			if err == nil {
				err = fmt.Errorf("drop database %q: %w", dbname, dropError)
			} else {
				err = fmt.Errorf("drop database %q: %w; previous error: %w", dbname, dropError, err)
			}
		}
	}()

	pool, err := pgxpool.Connect(ctx, i.connString+" dbname="+dbname)
	if err != nil {
		return fmt.Errorf("connect to database %q: %w", dbname, err)
	}
	defer pool.Close()

	if err = fn(pool); err != nil {
		return fmt.Errorf("in function: %w", err)
	}

	return nil
}

func (i *Instance) WithDatabaseSchema(ctx context.Context, schemaSQL string, fn func(pool *pgxpool.Pool) error) error {
	return i.WithDatabase(ctx, func(pool *pgxpool.Pool) error {
		_, err := pool.Exec(ctx, schemaSQL)
		if err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
		return fn(pool)
	})
}
