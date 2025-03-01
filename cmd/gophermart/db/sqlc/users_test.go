package db

import (
	"fmt"
	// "log"
	// "os"
	"testing"
    "github.com/jackc/pgx/v5"
)

// const (
// 	dbDriver = "postgres"
// 	dbSource = "postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable"
// )

// var ctx =context.Background()

// var testQueries *Queries

// func TestMain(m *testing.M) {
// 	conn, err := pgx.Connect(ctx, dbSource)
// 	if err != nil {
// 		log.Fatal("can not connect to db")
// 	}
// 	defer conn.Close(ctx)

// 	testQueries = New(conn)

// 	os.Exit(m.Run())

// }

func TestCreateUser(t *testing.T) {
	fmt.Println("Run test")
}