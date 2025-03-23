package main

import (
	"api"
	db "db/sqlc"
	"utils"
	"context"
	"log"
	"service"
	"time"
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



	// Создаем и запускаем worker
	worker := NewWorker(pool)
	go worker.Start(5 * time.Second) // Обновление каждые 5 секунд

	err = server.Start(config.ServerAddress) //serverAddress
	if err != nil {
		log.Fatal("Can not start server", err)
	}

}

type Worker struct {
	db *pgxpool.Pool
}

func NewWorker(db *pgxpool.Pool) *Worker {
	return &Worker{db: db}
}

// Start запускает worker, который обновляет баланс каждые N минут
func (w *Worker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.UpdateBalances()
		}
		}
}

// UpdateBalances обновляет балансы всех пользователей
func (w *Worker) UpdateBalances() {
	ctx := context.Background()

	// Получаем список всех пользователей
	rows, err := w.db.Query(ctx, "SELECT user_id FROM loyalty_accounts")
	if err != nil {
		log.Printf("Ошибка при получении списка пользователей: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Ошибка при сканировании user_id: %v", err)
			continue
		}

		// Вычисляем новый баланс (например, из внешней системы)
		newBalance, err := w.CalculateNewBalance(userID)
		if err != nil {
			log.Printf("Ошибка при расчете баланса для пользователя %s: %v", userID, err)
			continue
		}

		// Обновляем баланс в базе данных
		_, err = w.db.Exec(ctx, `
			UPDATE loyalty_accounts
			SET current_balance = $1, last_updated = $2
			WHERE user_id = $3
		`, newBalance, time.Now(), userID)
		if err != nil {
			log.Printf("Ошибка при обновлении баланса для пользователя %s: %v", userID, err)
			continue
		}

		log.Printf("Баланс пользователя %s обновлен: %f", userID, newBalance)
	}
}


// CalculateNewBalance вычисляет новый баланс для пользователя
func (w *Worker) CalculateNewBalance(userID string) (float64, error) {
	// Здесь можно реализовать логику расчета баланса
	// Например, запрос к внешнему API или вычисление на основе данных в базе
	return 1000.0, nil // Пример: возвращаем фиксированное значение
}