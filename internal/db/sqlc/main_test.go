package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:P@ssw0rd@localhost:5432/gophermart?sslmode=disable" 
)



var ctx = context.Background()
var testQueries *Queries
var testDB      *pgxpool.Pool

func TestMain(m *testing.M) {
	conn, err := pgx.Connect(ctx, dbSource)
	if err != nil {
		log.Fatal("ca not connect to db", err)
	}
	defer conn.Close(ctx)
	config, err := pgxpool.ParseConfig(dbSource)

	testDB, err = pgxpool.NewWithConfig(ctx, config)
	defer testDB.Close()

	testQueries = New(conn)

	os.Exit(m.Run())
}