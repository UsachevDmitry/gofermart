package main

import (
	"api"
	db "db/sqlc"
	"utils"
	"context"
	"log"
	"service"
	//"time"
	"github.com/jackc/pgx/v5/pgxpool"
)

// const (
// 	dbSource = "postgresql://postgres:P@ssw0rd@localhost:5432/gophermart?sslmode=disable"
// 	serverAddress = "127.0.0.1:8000"
// )

func main() {
	// Загрузка конфигурации
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Применение миграций
	utils.RunMigrations(config.DBSource)

	// Вывод загруженной конфигурации
	log.Printf("Конфигурация загружена: %+v\n", config)

	// Использование конфигурации
	log.Printf("Адрес сервера: %s\n", config.ServerAddress)
	log.Printf("Адрес базы данных: %s\n", config.DBSource)
	log.Printf("Адрес системы расчёта начислений: %s\n", config.AccrualSystemAddress)

	pool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}

	defer pool.Close()

	store := db.NewStore(pool)
	// Инициализация сервиса
	orderService := service.NewOrderService(store, &config)
	server := api.NewServer(store, orderService)



    // // Запуск воркера для обновления статусов заказов
    // go StartOrderStatusUpdater(10 * time.Second)

	err = server.Start(config.ServerAddress) //serverAddress
	if err != nil {
		log.Fatal("Can not start server", err)
	}

}
