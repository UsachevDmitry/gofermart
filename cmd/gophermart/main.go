package main

import (
	"api"
	db "db/sqlc"
	"utils"
	"context"
	"log"
	"service"
	// "time"
	"github.com/jackc/pgx/v5/pgxpool"
	// "fmt"
	// "net/http"
	// "errors"
	// "encoding/json"
	// "io"
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



	// // Создаем и запускаем worker
	// worker := NewWorker(pool)
	// go worker.Start(1 * time.Second) // Обновление каждые 1 секунду

	err = server.Start(config.ServerAddress) //serverAddress
	if err != nil {
		log.Fatal("Can not start server", err)
	}

}

// type Worker struct {
// 	db *pgxpool.Pool
// }

// func NewWorker(db *pgxpool.Pool) *Worker {
// 	return &Worker{db: db}
// }

// // Start запускает worker, который обновляет баланс каждые N минут
// func (w *Worker) Start(interval time.Duration) {
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			w.UpdateBalances()
// 		}
// 		}
// }

// // UpdateBalances обновляет балансы всех пользователей
// func (w *Worker) UpdateBalances() {
// 	ctx := context.Background()

// 	// Получаем список всех пользователей
// 	rows, err := w.db.Query(ctx, "SELECT user_id FROM loyalty_accounts")
// 	if err != nil {
// 		log.Printf("Ошибка при получении списка пользователей: %v", err)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var userID string
// 		if err := rows.Scan(&userID); err != nil {
// 			log.Printf("Ошибка при сканировании user_id: %v", err)
// 			continue
// 		}

// 		// Вычисляем новый баланс (например, из внешней системы)
// 		newBalance, err := w.CalculateNewBalance(userID)
// 		if err != nil {
// 			log.Printf("Ошибка при расчете баланса для пользователя %s: %v", userID, err)
// 			continue
// 		}

// 		// Обновляем баланс в базе данных
// 		_, err = w.db.Exec(ctx, `
// 			UPDATE loyalty_accounts
// 			SET current_balance = $1, last_updated = $2
// 			WHERE user_id = $3
// 		`, newBalance, time.Now(), userID)
// 		if err != nil {
// 			log.Printf("Ошибка при обновлении баланса для пользователя %s: %v", userID, err)
// 			continue
// 		}

// 		log.Printf("Баланс пользователя %s обновлен: %f", userID, newBalance)
// 	}
// }


// // CalculateNewBalance вычисляет новый баланс для пользователя
// func (w *Worker) CalculateNewBalance(userID string) (float64, error) {
// 	// Получаем список заказов пользователя из базы данных
// 	ctx := context.Background()
// 	rows, err := w.db.Query(ctx, `
// 		SELECT order_number
// 		FROM orders
// 		WHERE user_id = $1 AND status != 'PROCESSED' AND status != 'INVALID'
// 	`, userID)
// 	if err != nil {
// 		return 0, fmt.Errorf("ошибка при получении заказов: %v", err)
// 	}
// 	defer rows.Close()

// 	var totalAccrual float64

// 	// Обрабатываем каждый заказ
// 	for rows.Next() {
// 		var orderNumber string
// 		if err := rows.Scan(&orderNumber); err != nil {
// 			return 0, fmt.Errorf("ошибка при сканировании номера заказа: %v", err)
// 		}

// 		// Запрашиваем информацию о начислении для заказа
// 		accrual, err := w.getAccrualForOrder(orderNumber)
// 		if err != nil {
// 			return 0, fmt.Errorf("ошибка при запросе начисления для заказа %s: %v", orderNumber, err)
// 		}

// 		// Суммируем начисления
// 		totalAccrual += accrual
// 	}

// 	return totalAccrual, nil
// }

// type AccrualResponse struct {
// 	Order   string  `json:"order"`
// 	Status  string  `json:"status"`
// 	Accrual float64 `json:"accrual,omitempty"` // omitempty, так как поле может отсутствовать
// }

// func (w *Worker) getAccrualForOrder(orderNumber string) (float64, error) {
// 	config, err := utils.LoadConfig(".")
// 	if err != nil {
// 		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
// 	}
// 	// Формируем URL для запроса
// 	url := fmt.Sprintf("http://%s/api/orders/%s", config.AccrualSystemAddress, orderNumber)

// 	// Выполняем GET-запрос
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return 0, fmt.Errorf("ошибка при выполнении запроса: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Обрабатываем возможные коды ответа
// 	switch resp.StatusCode {
// 	case http.StatusOK:
// 		// Парсим ответ
// 		var accrualResp AccrualResponse
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			return 0, fmt.Errorf("ошибка при чтении ответа: %v", err)
// 		}

// 		if err := json.Unmarshal(body, &accrualResp); err != nil {
// 			return 0, fmt.Errorf("ошибка при парсинге JSON: %v", err)
// 		}

// 		// Проверяем статус заказа
// 		switch accrualResp.Status {
// 		case "PROCESSED":
// 			return accrualResp.Accrual, nil
// 		case "INVALID", "REGISTERED", "PROCESSING":
// 			return 0, nil
// 		default:
// 			return 0, fmt.Errorf("неизвестный статус заказа: %s", accrualResp.Status)
// 		}

// 	case http.StatusNoContent:
// 		// Заказ не зарегистрирован в системе
// 		return 0, nil

// 	case http.StatusTooManyRequests:
// 		// Превышено количество запросов
// 		retryAfter := resp.Header.Get("Retry-After")
// 		if retryAfter == "" {
// 			retryAfter = "60" // Значение по умолчанию
// 		}

// 		retryDuration, err := time.ParseDuration(retryAfter + "s")
// 		if err != nil {
// 			return 0, fmt.Errorf("ошибка при парсинге Retry-After: %v", err)
// 		}

// 		// Ждем и повторяем запрос
// 		time.Sleep(retryDuration)
// 		return w.getAccrualForOrder(orderNumber)

// 	case http.StatusInternalServerError:
// 		// Внутренняя ошибка сервера
// 		return 0, errors.New("внутренняя ошибка сервера системы начислений")

// 	default:
// 		return 0, fmt.Errorf("неожиданный статус ответа: %s", resp.Status)
// 	}
// }
