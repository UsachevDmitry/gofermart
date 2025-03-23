// package api

// import (
//     "net/http"
//     "github.com/gin-gonic/gin"
// 	"github.com/jackc/pgx/v5/pgtype"
// 	"utils"
// 	"time"
//     "errors"
//     "database/sql"
//     "log"
//     "math/big"
// 	db "db/sqlc"
// )

// type BalanceResponse struct {
//     Current   pgtype.Numeric `json:"current"`
//     Withdrawn pgtype.Numeric `json:"withdrawn"`
// }

// type WithdrawRequest struct {
//     Order string  `json:"order"`
//     Sum   pgtype.Numeric `json:"sum"`
// }


// // func (s *Server) getBalance(ctx *gin.Context) {
// //     // Проверка аутентификации
// //     login, exists := ctx.Get("login")
// //     if !exists {
// //         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
// //         return
// //     }

// //     // Получаем пользователя
// //     user, err := s.store.GetUser(ctx, login.(string))
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }

// //     // Преобразование userID в pgtype.Int4
// //     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

// //     // Получаем данные о балансе пользователя
// //     balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }

// //     // Форматируем ответ
// //     response := BalanceResponse{
// //         Current:   balance.CurrentBalance,
// //         Withdrawn: balance.WithdrawnBalance,
// //     }

// //     // Возвращаем ответ
// //     ctx.JSON(http.StatusOK, response) // 200
// // }

// func (s *Server) getBalance(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }

//     // Получаем пользователя
//     user, err := s.store.GetUser(ctx, login.(string))
//     if err != nil {
//         if errors.Is(err, sql.ErrNoRows) {
//             ctx.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"}) // 404
//         } else {
//             log.Printf("Ошибка при получении пользователя: %v", err)
//             ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         }
//         return
//     }

//     // Проверка user.ID
//     if user.ID == 0 {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "неверный ID пользователя"}) // 500
//         return
//     }

//     // Преобразование userID в pgtype.Int4
//     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

//     // Получаем данные о балансе пользователя
//     balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         if errors.Is(err, sql.ErrNoRows) {
//             // Если баланс не найден, возвращаем нулевой баланс
//             zeroBalance := pgtype.Numeric{
//                 Int:    big.NewInt(0), // Значение 0
//                 Exp:    0,             // Экспонента (0 для целых чисел)
//                 Valid:  true,          // Указываем, что значение валидно
//             }

//             response := BalanceResponse{
//                 Current:   zeroBalance,
//                 Withdrawn: zeroBalance,
//             }
//             // response := BalanceResponse{
//             //     Current: 0,
//             //     Withdrawn: 0,
//             // }
//             ctx.JSON(http.StatusOK, response) // 200
//             return
//         } else {
//             log.Printf("Ошибка при получении баланса: %v", err)
//             ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//             return
//         }
//     }

//     // Форматируем ответ
//     response := BalanceResponse{
//         Current:   balance.CurrentBalance,
//         Withdrawn: balance.WithdrawnBalance,
//     }

//     // Возвращаем ответ
//     ctx.JSON(http.StatusOK, response) // 200
// }

// // func (s *Server) withdrawBalance(ctx *gin.Context) {
// //     // Проверка аутентификации
// //     login, exists := ctx.Get("login")
// //     if !exists {
// //         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
// //         return
// //     }
// //     user, err := s.store.GetUser(ctx, login.(string))
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }

// //     // Преобразование userID в pgtype.Int4
// //     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

// //     // Парсим тело запроса
// //     var req WithdrawRequest
// //     if err := ctx.ShouldBindJSON(&req); err != nil {
// //         ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
// //         return
// //     }

// //     // Проверяем номер заказа с помощью алгоритма Луна
// //     if !utils.IsValidLuhn(req.Order) {
// //         ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "неверный номер заказа"}) // 422
// //         return
// //     }

// //     // Проверяем, достаточно ли средств на счету
// //     balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }
// //     if balance.CurrentBalance < req.Sum {
// //         ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "на счету недостаточно средств"}) // 402
// //         return
// //     }

// //     // Регистрируем списание
// //     err = s.store.CreateWithdrawal(ctx.Request.Context(), db.CreateWithdrawalParams{
// //         UserID:      userIDInt4,
// //         OrderNumber: req.Order,
// //         Sum:         req.Sum,
// //     })
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }

// //     // Обновляем баланс пользователя
// //     err = s.store.UpdateBalance(ctx.Request.Context(), db.UpdateBalanceParams{
// //         UserID:           userIDInt4,
// //         CurrentBalance:   balance.CurrentBalance - req.Sum,
// //         WithdrawnBalance: balance.WithdrawnBalance + req.Sum,
// //     })
// //     if err != nil {
// //         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
// //         return
// //     }

// //     // Успешный ответ
// //     ctx.Status(http.StatusOK) // 200
// // }

// func (s *Server) withdrawBalance(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
//     user, err := s.store.GetUser(ctx, login.(string))
//     if err != nil {
//         log.Printf("Failed to get user: %v", err)
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Преобразование userID в pgtype.Int4
//     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

//     // Парсим тело запроса
//     var req WithdrawRequest
//     if err := ctx.ShouldBindJSON(&req); err != nil {
//         ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
//         return
//     }

//     // Проверяем номер заказа с помощью алгоритма Луна
//     if !utils.IsValidLuhn(req.Order) {
//         ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "неверный номер заказа"}) // 422
//         return
//     }

//     // Проверяем, достаточно ли средств на счету
//     balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         log.Printf("Failed to get balance: %v", err)
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }
//     if balance.CurrentBalance < req.Sum {
//         ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "на счету недостаточно средств"}) // 402
//         return
//     }

//     // Регистрируем списание
//     err = s.store.CreateWithdrawal(ctx.Request.Context(), db.CreateWithdrawalParams{
//         UserID:      userIDInt4,
//         OrderNumber: req.Order,
//         Sum:         req.Sum,
//     })
//     if err != nil {
//         log.Printf("Failed to create withdrawal: %v", err)
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Обновляем баланс пользователя
//     err = s.store.UpdateBalance(ctx.Request.Context(), db.UpdateBalanceParams{
//         UserID:           userIDInt4,
//         CurrentBalance:   balance.CurrentBalance - req.Sum,
//         WithdrawnBalance: balance.WithdrawnBalance + req.Sum,
//     })
//     if err != nil {
//         log.Printf("Failed to update balance: %v", err)
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Успешный ответ
//     ctx.Status(http.StatusOK) // 200
// }

// type WithdrawalResponse struct {
//     Order       string    `json:"order"`
//     Sum         float64   `json:"sum"`
//     ProcessedAt time.Time `json:"processed_at"`
// }

// func (s *Server) getWithdrawals(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
//     user, err := s.store.GetUser(ctx, login.(string))
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Преобразование userID в pgtype.Int4
//     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

//     // Получаем список списаний пользователя
//     withdrawals, err := s.store.GetWithdrawalsByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Если списаний нет, возвращаем 204
//     if len(withdrawals) == 0 {
//         ctx.Status(http.StatusNoContent) // 204
//         return
//     }

//     // Форматируем ответ
//     var response []WithdrawalResponse
//     for _, withdrawal := range withdrawals {
//         response = append(response, WithdrawalResponse{
//             Order:       withdrawal.OrderNumber,
//             Sum:         withdrawal.Sum,
//             ProcessedAt: withdrawal.ProcessedAt.Time,
//         })
//     }

//     // Возвращаем ответ
//     ctx.JSON(http.StatusOK, response) // 200
// }

package api

import (
    "database/sql"
    "errors"
    "log"
    "math/big"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgtype"
    db "db/sqlc"
    "utils"
)

type BalanceResponse struct {
    Current   pgtype.Numeric `json:"current"`
    Withdrawn pgtype.Numeric `json:"withdrawn"`
}

type WithdrawRequest struct {
    Order string         `json:"order"`
    Sum   pgtype.Numeric `json:"sum"`
}

type WithdrawalResponse struct {
    Order       string         `json:"order"`
    Sum         pgtype.Numeric `json:"sum"`
    ProcessedAt time.Time      `json:"processed_at"`
}

// Преобразует pgtype.Numeric в float64
func numericToFloat64(n pgtype.Numeric) (float64, error) {
    if !n.Valid {
        return 0, errors.New("значение не валидно")
    }

    // Преобразуем pgtype.Numeric в pgtype.Float8
    float8, err := n.Float64Value()
    if err != nil {
        return 0, err
    }

    // Возвращаем значение float64
    return float8.Float64, nil
}

// Преобразует float64 в pgtype.Numeric
func float64ToNumeric(f float64) (pgtype.Numeric, error) {
    var numeric pgtype.Numeric
    err := numeric.Scan(f)
    return numeric, err
}

func (s *Server) getBalance(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }

    // Получаем пользователя
    user, err := s.store.GetUser(ctx, login.(string))
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"}) // 404
        } else {
            log.Printf("Ошибка при получении пользователя: %v", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        }
        return
    }

    // Проверка user.ID
    if user.ID == 0 {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "неверный ID пользователя"}) // 500
        return
    }

    // Преобразование userID в pgtype.Int4
    userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

    // Получаем данные о балансе пользователя
    balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // Если баланс не найден, возвращаем нулевой баланс
            zeroBalance := pgtype.Numeric{
                Int:    big.NewInt(0), // Значение 0
                Exp:    0,             // Экспонента (0 для целых чисел)
                Valid:  true,          // Указываем, что значение валидно
            }

            response := BalanceResponse{
                Current:   zeroBalance,
                Withdrawn: zeroBalance,
            }
            ctx.JSON(http.StatusOK, response) // 200
            return
        } else {
            log.Printf("Ошибка при получении баланса: %v", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
            return
        }
    }

    // Форматируем ответ
    response := BalanceResponse{
        Current:   balance.CurrentBalance,
        Withdrawn: balance.WithdrawnBalance,
    }

    // Возвращаем ответ
    ctx.JSON(http.StatusOK, response) // 200
}

func (s *Server) withdrawBalance(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }
    user, err := s.store.GetUser(ctx, login.(string))
    if err != nil {
        log.Printf("Failed to get user: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Преобразование userID в pgtype.Int4
    userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

    // Парсим тело запроса
    var req WithdrawRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
        return
    }

    // Проверяем номер заказа с помощью алгоритма Луна
    if !utils.IsValidLuhn(req.Order) {
        ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "неверный номер заказа"}) // 422
        return
    }

    // Проверяем, достаточно ли средств на счету
    balance, err := s.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        log.Printf("Failed to get balance: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Преобразуем pgtype.Numeric в float64 для сравнения
    currentBalance, err := numericToFloat64(balance.CurrentBalance)
    if err != nil {
        log.Printf("Failed to convert current balance: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    reqSum, err := numericToFloat64(req.Sum)
    if err != nil {
        log.Printf("Failed to convert request sum: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    if currentBalance < reqSum {
        ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "на счету недостаточно средств"}) // 402
        return
    }

    // Регистрируем списание
    err = s.store.CreateWithdrawal(ctx.Request.Context(), db.CreateWithdrawalParams{
        UserID:      userIDInt4,
        OrderNumber: req.Order,
        Sum:         reqSum,
    })
    if err != nil {
        log.Printf("Failed to create withdrawal: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Обновляем баланс пользователя
    newCurrentBalance, err := float64ToNumeric(currentBalance - reqSum)
    if err != nil {
        log.Printf("Failed to convert new current balance: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }
    withdrawnBalance, err := numericToFloat64(balance.WithdrawnBalance)
    if err != nil {
        log.Printf("Failed to convert request sum: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }
    newWithdrawnBalance, err := float64ToNumeric(withdrawnBalance + reqSum)
    if err != nil {
        log.Printf("Failed to convert new withdrawn balance: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    err = s.store.UpdateBalance(ctx.Request.Context(), db.UpdateBalanceParams{
        UserID:           userIDInt4,
        CurrentBalance:   newCurrentBalance,
        WithdrawnBalance: newWithdrawnBalance,
    })
    if err != nil {
        log.Printf("Failed to update balance: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Успешный ответ
    ctx.Status(http.StatusOK) // 200
}

func (s *Server) getWithdrawals(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }
    user, err := s.store.GetUser(ctx, login.(string))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Преобразование userID в pgtype.Int4
    userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

    // Получаем список списаний пользователя
    withdrawals, err := s.store.GetWithdrawalsByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Если списаний нет, возвращаем 204
    if len(withdrawals) == 0 {
        ctx.Status(http.StatusNoContent) // 204
        return
    }

// Форматируем ответ
var response []WithdrawalResponse
for _, withdrawal := range withdrawals {
    // Преобразуем withdrawal.Sum (float64) в pgtype.Numeric
    sumNumeric, err := float64ToNumeric(withdrawal.Sum)
    if err != nil {
        log.Printf("Failed to convert withdrawal sum: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    response = append(response, WithdrawalResponse{
        Order:       withdrawal.OrderNumber,
        Sum:         sumNumeric, // Используем преобразованное значение
        ProcessedAt: withdrawal.ProcessedAt.Time,
    })
}

    // Возвращаем ответ
    ctx.JSON(http.StatusOK, response) // 200
}