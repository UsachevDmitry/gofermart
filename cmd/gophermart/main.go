// package main

// import (
// 	"api"
// 	db "db/sqlc"
// 	"utils"
// 	"context"
// 	"log"
// 	"service"
// 	"time"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"fmt"
// 	"net/http"
// 	"errors"
// 	"encoding/json"
// 	"io"
// )

// // const (
// // 	dbSource = "postgresql://postgres:P@ssw0rd@localhost:5432/gophermart?sslmode=disable"
// // 	serverAddress = "127.0.0.1:8000"
// // )

// func main() {
// 	// Загрузка конфигурации
// 	config, err := utils.LoadConfig(".")
// 	if err != nil {
// 		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
// 	}

// 	// Применение миграций
// 	utils.RunMigrations(config.DBSource)

// 	// Вывод загруженной конфигурации
// 	log.Printf("Конфигурация загружена: %+v\n", config)

// 	// Использование конфигурации
// 	log.Printf("Адрес сервера: %s\n", config.ServerAddress)
// 	log.Printf("Адрес базы данных: %s\n", config.DBSource)
// 	log.Printf("Адрес системы расчёта начислений: %s\n", config.AccrualSystemAddress)

// 	pool, err := pgxpool.New(context.Background(), config.DBSource)
// 	if err != nil {
// 		log.Fatal("can not connect to db", err)
// 	}

// 	defer pool.Close()

// 	store := db.NewStore(pool)
// 	// Инициализация сервиса
// 	orderService := service.NewOrderService(store, &config)
// 	server := api.NewServer(store, orderService)



// 	// Создаем и запускаем worker
// 	worker := NewWorker(pool)
// 	go worker.Start(3 * time.Second) // Обновление каждые 3 секунду

// 	err = server.Start(config.ServerAddress) //serverAddress
// 	if err != nil {
// 		log.Fatal("Can not start server", err)
// 	}

// }

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

	  
		  
// 	for range ticker.C{
// 			w.UpdateBalances()
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
	"fmt"
	"net/http"
	"errors"
	"encoding/json"
	"io"
	"os"
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

	worker := NewWorker(pool)
	
	// Настройка интервала для тестового режима
	if os.Getenv("TEST_MODE") == "true" {
		go worker.Start(5 * time.Second)
	} else {
		go worker.Start(3 * time.Second)
	}

	err = server.Start(config.ServerAddress)
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

func (w *Worker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		w.UpdateBalances()
	}
}

func (w *Worker) UpdateBalances() {
	ctx := context.Background()

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

		newBalance, err := w.CalculateNewBalance(userID)
		if err != nil {
			log.Printf("Ошибка при расчете баланса для пользователя %s: %v", userID, err)
			continue
		}

		_, err = w.db.Exec(ctx, `
			UPDATE loyalty_accounts
			SET current_balance = $1, last_updated = $2
			WHERE user_id = $3
		`, newBalance, time.Now(), userID)
		if err != nil {
			log.Printf("Ошибка при обновлении баланса для пользователя %s: %v", userID, err)
		}
	}
}

func (w *Worker) CalculateNewBalance(userID string) (float64, error) {
	ctx := context.Background()
	rows, err := w.db.Query(ctx, `
		SELECT order_number
		FROM orders
		WHERE user_id = $1 AND status IN ('NEW', 'PROCESSING')
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении заказов: %v", err)
	}
	defer rows.Close()

	var totalAccrual float64

	for rows.Next() {
		var orderNumber string
		if err := rows.Scan(&orderNumber); err != nil {
			return 0, fmt.Errorf("ошибка при сканировании номера заказа: %v", err)
		}

		accrual, err := w.getAccrualForOrder(orderNumber)
		if err != nil {
			var tooManyErr *ErrTooManyRequests
			if errors.As(err, &tooManyErr) {
				return 0, err
			}
			continue
		}

		totalAccrual += accrual
	}

	return totalAccrual, nil
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (w *Worker) getAccrualForOrder(orderNumber string) (float64, error) {
	config, err := utils.LoadConfig(".")
	if err != nil {
		return 0, fmt.Errorf("ошибка загрузки конфигурации: %v", err)
	}

	url := fmt.Sprintf("http://%s/api/orders/%s", config.AccrualSystemAddress, orderNumber)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("ошибка при чтении ответа: %v", err)
		}

		if err := json.Unmarshal(body, &accrualResp); err != nil {
			return 0, fmt.Errorf("ошибка при парсинге JSON: %v", err)
		}

		// Обновляем статус заказа в базе данных
		if err := w.updateOrderStatus(orderNumber, accrualResp.Status); err != nil {
			return 0, err
		}

		if accrualResp.Status == "PROCESSED" {
			return accrualResp.Accrual, nil
		}
		return 0, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter == "" {
			retryAfter = "60"
		}

		retryDuration, err := time.ParseDuration(retryAfter + "s")
		if err != nil {
			return 0, fmt.Errorf("ошибка при парсинге Retry-After: %v", err)
		}

		return 0, &ErrTooManyRequests{RetryAfter: retryDuration}

	default:
		return 0, fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}
}

// Добавляем маппинг статусов
func mapStatus(accrualStatus string) string {
	switch accrualStatus {
	case "REGISTERED":
		return "NEW"
	case "PROCESSING":
		return "PROCESSING"
	case "PROCESSED":
		return "PROCESSED"
	case "INVALID":
		return "INVALID"
	default:
		return "UNKNOWN"
	}
}

// Обновляем статус заказа в базе
func (w *Worker) updateOrderStatus(orderNumber, accrualStatus string) error {
	mappedStatus := mapStatus(accrualStatus)
	ctx := context.Background()

	_, err := w.db.Exec(ctx, `
		UPDATE orders 
		SET status = $1 
		WHERE order_number = $2 
		  AND status NOT IN ('PROCESSED', 'INVALID')
	`, mappedStatus, orderNumber)

	return err
}

type ErrTooManyRequests struct {
	RetryAfter time.Duration
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after %s", e.RetryAfter)
}