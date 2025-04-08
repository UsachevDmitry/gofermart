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
    "errors"
    "strconv"
    "strings"
    "sync"
    "log"
)

type AccrualResponse struct {
    Order   string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual,omitempty"`
}

// type OrderService struct {
//     repo *db.Store
//     config *utils.Config }

type OrderService struct {
    repo        *db.Store
    config      *utils.Config
    pendingJobs map[string]bool // Трекер активных заказов
    mu          sync.Mutex      // Для безопасного доступа к pendingJobs
}

// func NewOrderService(repo *db.Store, config *utils.Config) *OrderService {
//     return &OrderService{
//         repo:   repo,
//         config: config,
//     }
// }

func NewOrderService(repo *db.Store, config *utils.Config) *OrderService {
    s := &OrderService{
        repo:       repo,
        config:     config,
        pendingJobs: make(map[string]bool),
    }
    go s.restorePendingOrders() // Восстановление при старте
    return s
}

// func (s *OrderService) restorePendingOrders() {
//     ctx := context.Background()
//     orders, err := s.repo.GetUnprocessedOrders(ctx) // Нужно добавить этот метод в репозиторий
//     if err != nil {
//         log.Printf("failed to restore pending orders: %v", err)
//         return
//     }

//     for _, order := range orders {
//         s.mu.Lock()
//         if !s.pendingJobs[order.OrderNumber] {
//             s.pendingJobs[order.OrderNumber] = true
//             go s.PollOrderStatus(order.OrderNumber)
//         }
//         s.mu.Unlock()
//     }
// }

func (s *OrderService) restorePendingOrders() {
    ctx := context.Background()
    
    // Используем сгенерированный sqlc метод
    orders, err := s.repo.GetUnprocessedOrders(ctx)
    if err != nil {
        log.Printf("failed to get unprocessed orders: %v", err)
        return
    }
    
    for _, order := range orders {
        go s.PollOrderStatus(order.OrderNumber)
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
        newAccrual, _ := float64ToNumeric(accrual)
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

// func (s *OrderService) PollOrderStatus(orderNumber string) error {
//     ctx := context.Background()
//     maxAttempts := 10
//     baseInterval := 2 * time.Second

//     for attempt := 0; attempt < maxAttempts; attempt++ {
//         status, accrual, err := s.fetchAccrualStatus(ctx, orderNumber)
//         if err != nil {
//             // Если получили ошибку с рекомендацией Retry-After
//             if retryAfter := parseRetryAfterError(err); retryAfter > 0 {
//                 time.Sleep(retryAfter)
//                 continue
//             }
//             return fmt.Errorf("failed to fetch accrual status: %w", err)
//         }

//         if err := s.UpdateOrder(ctx, orderNumber, status, accrual); err != nil {
//             return fmt.Errorf("failed to update order: %w", err)
//         }

//         if status == "PROCESSED" || status == "INVALID" {
//             return nil
//         }

//         time.Sleep(baseInterval)
//     }

//     return fmt.Errorf("failed to process order %s after %d attempts", orderNumber, maxAttempts)
// }

func (s *OrderService) PollOrderStatus(orderNumber string) error {
    s.mu.Lock()
    if s.pendingJobs[orderNumber] {
        s.mu.Unlock()
        return nil // Уже обрабатывается
    }
    s.pendingJobs[orderNumber] = true
    s.mu.Unlock()

    defer func() {
        s.mu.Lock()
        delete(s.pendingJobs, orderNumber)
        s.mu.Unlock()
    }()

    ctx := context.Background()
    maxAttempts := 10
    baseInterval := 2 * time.Second

    for attempt := 0; attempt < maxAttempts; attempt++ {
        status, accrual, err := s.fetchAccrualStatus(ctx, orderNumber)
        if err != nil {
            if retryAfter := parseRetryAfterError(err); retryAfter > 0 {
                time.Sleep(retryAfter)
                continue
            }
            return fmt.Errorf("failed to fetch accrual status: %w", err)
        }

        if err := s.UpdateOrder(ctx, orderNumber, status, accrual); err != nil {
            return fmt.Errorf("failed to update order: %w", err)
        }

        if status == "PROCESSED" || status == "INVALID" {
            return nil
        }

        time.Sleep(baseInterval)
    }

    return fmt.Errorf("failed to process order %s after %d attempts", orderNumber, maxAttempts)
}

func parseRetryAfterError(err error) time.Duration {
    const prefix = "rate limit exceeded, retry after "
    errStr := err.Error()
    
    if strings.HasPrefix(errStr, prefix) {
        secondsStr := strings.TrimPrefix(errStr, prefix)
        seconds, err := strconv.Atoi(secondsStr)
        if err == nil {
            return time.Duration(seconds) * time.Second
        }
    }
    return 0
}

func (s *OrderService) fetchAccrualStatus(ctx context.Context, orderNumber string) (string, float64, error) {
    url := fmt.Sprintf("%s/api/orders/%s", s.config.AccrualSystemAddress, orderNumber)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return "", 0, fmt.Errorf("failed to create request: %w", err)
    }

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", 0, fmt.Errorf("failed to make request to accrual system: %w", err)
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusOK:
        var accrualResp AccrualResponse
        if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
            return "", 0, fmt.Errorf("failed to decode response: %w", err)
        }
        return accrualResp.Status, accrualResp.Accrual, nil

    case http.StatusNoContent:
        return "REGISTERED", 0, nil

    case http.StatusTooManyRequests:
        retryAfter := resp.Header.Get("Retry-After")
        if retryAfter == "" {
            retryAfter = "60"
        }
        return "", 0, fmt.Errorf("rate limit exceeded, retry after %s", retryAfter)

    case http.StatusInternalServerError:
        return "", 0, fmt.Errorf("internal server error in accrual system")

    default:
        return "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
}

func (s *OrderService) GetUserBalance(ctx context.Context, userID int32) (current, withdrawn float64, err error) {
    // Получаем все PROCESSED заказы пользователя
    processedOrders, err := s.repo.GetProcessedOrdersByUserID(ctx, pgtype.Int4{Int32: userID, Valid: true})
    if err != nil {
        return 1, 1, fmt.Errorf("failed to get processed orders: %w", err)
    }

    // Считаем общее начисление
    var totalAccrual float64
    for _, order := range processedOrders {
        if order.Accrual.Valid {
            totalAccrual += order.Accrual.Float64
        }
    }

    // Получаем все списания пользователя
    withdrawals, err := s.repo.GetWithdrawalsByUserID(ctx, pgtype.Int4{Int32: userID, Valid: true})
    if err != nil {
        return 2, 2, fmt.Errorf("failed to get withdrawals: %w", err)
    }

    // Считаем общую сумму списаний
    var totalWithdrawn float64
    for _, w := range withdrawals {
        totalWithdrawn += w.Sum
    }

    return totalAccrual - totalWithdrawn, totalWithdrawn, nil
}

func (s *OrderService) Withdraw(ctx context.Context, userID int32, orderNumber string, sum float64) error {
    // Проверяем баланс
    current, _, err := s.GetUserBalance(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to check balance: %w", err)
    }

    if current < sum {
        return errors.New("insufficient funds")
    }

    // Создаем запись о списании
    err = s.repo.CreateWithdrawal(ctx, db.CreateWithdrawalParams{
        UserID:      pgtype.Int4{Int32: userID, Valid: true},
        OrderNumber: orderNumber,
        Sum:         sum,
    })
    if err != nil {
        return fmt.Errorf("failed to create withdrawal: %w", err)
    }

    return nil
}

func (s *OrderService) GetUserWithdrawals(ctx context.Context, userID int32) ([]db.Withdrawal, error) {
    withdrawals, err := s.repo.GetWithdrawalsByUserID(ctx, pgtype.Int4{Int32: userID, Valid: true})
    if err != nil {
        return nil, fmt.Errorf("failed to get withdrawals: %w", err)
    }
    
    var result []db.Withdrawal
    for _, w := range withdrawals {
        result = append(result, db.Withdrawal{
            OrderNumber: w.OrderNumber,
            UserID:      w.UserID,
            Sum:         w.Sum,
            ProcessedAt: w.ProcessedAt,
        })
    }
    
    return result, nil
}
