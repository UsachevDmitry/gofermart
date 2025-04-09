package main

import (
	"api"
	"context"
	db "db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"service"
	"utils"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	utils.RunMigrations(config.DBSource)

	pool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)
	orderService := service.NewOrderService(store, &config)
	server := api.NewServer(store, orderService)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Can not start server", err)
	}
}
