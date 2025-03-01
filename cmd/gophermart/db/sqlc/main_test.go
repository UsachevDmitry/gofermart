package db

import (
	"context"
	"log"
	"os"
	"testing"
	"github.com/UsachevDmitry/gofermart/cmd/gophermart/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testDB *pgxpool.Pool
var testQueries *Queries

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("can not config file", err)
	}
	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	defer testDB.Close()

	testQueries = New(testDB)

	os.Exit(m.Run())
}
