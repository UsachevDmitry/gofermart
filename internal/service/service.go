package service

import (
    "context"
    "fmt"
    db "db/sqlc"
	"time"
    "github.com/jackc/pgx/v5/pgtype"
    "encoding/json"
    "net/http"
    "utils"
)

type OrderService struct {
    repo *db.Store
    config *utils.Config }

    func NewOrderService(repo *db.Store, config *utils.Config) *OrderService {
        return &OrderService{
            repo:   repo,
            config: config,
        }
    }

func float64ToNumeric(f float64) (pgtype.Numeric, error) {
    var numeric pgtype.Numeric
    err := numeric.Scan(f)
    return numeric, err
}

func (s *OrderService) UpdateOrder(ctx context.Context, orderNumber string, status string, accrual float64) error {
    // Обновляем статус и начисление заказа
    if err := s.repo.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
        Status:      status,
        Accrual:     pgtype.Float8{Float64: accrual, Valid: true},
        OrderNumber: orderNumber,
    }); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // Если заказ обработан, обновляем баланс пользователя
    if status == "PROCESSED" {
        // Получаем ID пользователя, которому принадлежит заказ
        userID, err := s.repo.GetOrderOwner(ctx, orderNumber)
        if err != nil {
            return fmt.Errorf("failed to get order owner: %w", err)
        }

        // Обновляем баланс пользователя
        newAccrual, err := float64ToNumeric(accrual)
        // if err != nil {
        //     log.Printf("Failed to convert new current balance: %v", err)
        //     ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        //     return 
        // }
        if err := s.repo.UpdateLoyaltyAccountBalance(ctx, db.UpdateLoyaltyAccountBalanceParams{
            CurrentBalance: newAccrual,
            UserID:         userID,
        }); err != nil {
            return fmt.Errorf("failed to update loyalty account balance: %w", err)
        }

        // Получаем ID аккаунта пользователя
        accountID, err := s.repo.GetLoyaltyAccountID(ctx, userID)
        if err != nil {
            return fmt.Errorf("failed to get loyalty account ID: %w", err)
        }

        // Получаем ID заказа
        orderID, err := s.repo.GetOrderID(ctx, orderNumber)
        if err != nil {
            return fmt.Errorf("failed to get order ID: %w", err)
        }

        // Создаем запись о транзакции
        if err := s.repo.CreateLoyaltyTransaction(ctx, db.CreateLoyaltyTransactionParams{
            AccountID:       pgtype.Int4{Int32: accountID, Valid: true},
            OrderID:         pgtype.Int4{Int32: orderID, Valid: true},
            Points:          accrual,
            TransactionType: "accrual",
        }); err != nil {
            return fmt.Errorf("failed to create loyalty transaction: %w", err)
        }
    }

    return nil
}

func (s *OrderService) PollOrderStatus(orderNumber string) error {
    ctx := context.Background() // Создаем контекст
    maxAttempts := 10           // Максимальное количество попыток
    interval := 2 * time.Second // Интервал между запросами

    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Запрашиваем статус заказа у внешней системы
        status, accrual, err := s.fetchAccrualStatus(ctx, orderNumber) // Передаем контекст
        if err != nil {
            return fmt.Errorf("failed to fetch accrual status: %w", err)
        }

        // Обновляем статус заказа в базе данных
        if err := s.UpdateOrder(ctx, orderNumber, status, accrual); err != nil { // Передаем контекст
            return fmt.Errorf("failed to update order: %w", err)
        }

        // Если заказ обработан, завершаем опрос
        if status == "PROCESSED" || status == "INVALID" {
            return nil
        }

        // Ждем перед следующим запросом
        time.Sleep(interval)
    }

    return fmt.Errorf("failed to process order %s after %d attempts", orderNumber, maxAttempts)
}

type AccrualResponse struct {
    Order   string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual,omitempty"`
}

func (s *OrderService) fetchAccrualStatus(ctx context.Context, orderNumber string) (string, float64, error) {
    // Формируем URL для запроса
    url := fmt.Sprintf("%s/api/orders/%s", s.config.AccrualSystemAddress, orderNumber)

    // Создаем HTTP-запрос с контекстом
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return "", 0, fmt.Errorf("failed to create request: %w", err)
    }

    // Создаем HTTP-клиент с таймаутом
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    // Выполняем GET-запрос
    resp, err := client.Do(req)
    if err != nil {
        return "", 0, fmt.Errorf("failed to make request to accrual system: %w", err)
    }
    defer resp.Body.Close()

    // Обрабатываем ответ
    switch resp.StatusCode {
    case http.StatusOK: // 200
        var accrualResp AccrualResponse
        if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
            return "", 0, fmt.Errorf("failed to decode response: %w", err)
        }
        return accrualResp.Status, accrualResp.Accrual, nil

    case http.StatusNoContent: // 204
        return "REGISTERED", 0, nil

    case http.StatusTooManyRequests: // 429
        retryAfter := resp.Header.Get("Retry-After")
        if retryAfter == "" {
            retryAfter = "60" // Значение по умолчанию, если заголовок отсутствует
        }
        return "", 0, fmt.Errorf("rate limit exceeded, retry after %s seconds", retryAfter)

    case http.StatusInternalServerError: // 500
        return "", 0, fmt.Errorf("internal server error in accrual system")

    default:
        return "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
}