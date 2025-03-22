package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"utils"
	"time"
	db "db/sqlc"
)

type BalanceResponse struct {
    Current   float64 `json:"current"`
    Withdrawn float64 `json:"withdrawn"`
}

// func (server *Server) getBalance(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
// 	user, _ := server.store.GetUser(ctx, login.(string))
// 	userID := user.ID
// 	// Преобразование userID в pgtype.Int4
// 	userIDInt4 := pgtype.Int4{Int32: userID, Valid: true}

//     // Получаем данные о балансе пользователя
//     balance, err := server.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Форматируем ответ
//     response := BalanceResponse{
//         Current:   balance.CurrentBalance,
//         Withdrawn: balance.WithdrawnBalance,
//     }

//     // Возвращаем ответ
//     ctx.JSON(http.StatusOK, response) // 200
// }

// func (server *Server) getBalance(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
//     user, err := server.store.GetUser(ctx, login.(string))
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Преобразование userID в pgtype.Int4
//     userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

//     // Получаем данные о балансе пользователя
//     balance, err := server.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Форматируем ответ
//     response := BalanceResponse{
//         Current:   balance.CurrentBalance,
//         Withdrawn: balance.WithdrawnBalance,
//     }

//     // Возвращаем ответ
//     ctx.JSON(http.StatusOK, response) // 200
// }

type WithdrawRequest struct {
    Order string  `json:"order"`
    Sum   float64 `json:"sum"`
}

func (server *Server) getBalance(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }

    // Получаем пользователя
    user, err := server.store.GetUser(ctx, login.(string))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Преобразование userID в pgtype.Int4
    userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

    // Получаем данные о балансе пользователя
    balance, err := server.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Форматируем ответ
    response := BalanceResponse{
        Current:   balance.CurrentBalance,
        Withdrawn: balance.WithdrawnBalance,
    }

    // Возвращаем ответ
    ctx.JSON(http.StatusOK, response) // 200
}

// func (server *Server) withdrawBalance(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
// 	user, _ := server.store.GetUser(ctx, login.(string))
// 	userID := user.ID
// 	// Преобразование userID в pgtype.Int4
// 	userIDInt4 := pgtype.Int4{Int32: userID, Valid: true}

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
//     balance, err := server.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }
//     if balance.CurrentBalance < req.Sum {
//         ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "на счету недостаточно средств"}) // 402
//         return
//     }

//     // Регистрируем списание
//     err = server.store.CreateWithdrawal(ctx.Request.Context(), db.CreateWithdrawalParams{
//         UserID:      userIDInt4,
//         OrderNumber: req.Order,
//         Sum:         req.Sum,
//     })
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Обновляем баланс пользователя
//     err = server.store.UpdateBalance(ctx.Request.Context(), db.UpdateBalanceParams{
//         UserID:           userIDInt4,
//         CurrentBalance:   balance.CurrentBalance - req.Sum,
//         WithdrawnBalance: balance.WithdrawnBalance + req.Sum,
//     })
//     if err != nil {
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Успешный ответ
//     ctx.Status(http.StatusOK) // 200
// }

func (server *Server) withdrawBalance(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }
    user, err := server.store.GetUser(ctx, login.(string))
    if err != nil {
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
    balance, err := server.store.GetBalanceByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }
    if balance.CurrentBalance < req.Sum {
        ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "на счету недостаточно средств"}) // 402
        return
    }

    // Регистрируем списание
    err = server.store.CreateWithdrawal(ctx.Request.Context(), db.CreateWithdrawalParams{
        UserID:      userIDInt4,
        OrderNumber: req.Order,
        Sum:         req.Sum,
    })
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Обновляем баланс пользователя
    err = server.store.UpdateBalance(ctx.Request.Context(), db.UpdateBalanceParams{
        UserID:           userIDInt4,
        CurrentBalance:   balance.CurrentBalance - req.Sum,
        WithdrawnBalance: balance.WithdrawnBalance + req.Sum,
    })
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Успешный ответ
    ctx.Status(http.StatusOK) // 200
}

type WithdrawalResponse struct {
    Order       string    `json:"order"`
    Sum         float64   `json:"sum"`
    ProcessedAt time.Time `json:"processed_at"`
}

// func (server *Server) getWithdrawals(ctx *gin.Context) {
//     // Проверка аутентификации
//     login, exists := ctx.Get("login")
//     if !exists {
//         ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
//         return
//     }
// 	user, _ := server.store.GetUser(ctx, login.(string))
// 	userID := user.ID
// 	// Преобразование userID в pgtype.Int4
// 	userIDInt4 := pgtype.Int4{Int32: userID, Valid: true}

//     // Получаем список списаний пользователя
//     withdrawals, err := server.store.GetWithdrawalsByUserID(ctx.Request.Context(), userIDInt4)
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

func (server *Server) getWithdrawals(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }
    user, err := server.store.GetUser(ctx, login.(string))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Преобразование userID в pgtype.Int4
    userIDInt4 := pgtype.Int4{Int32: user.ID, Valid: true}

    // Получаем список списаний пользователя
    withdrawals, err := server.store.GetWithdrawalsByUserID(ctx.Request.Context(), userIDInt4)
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
        response = append(response, WithdrawalResponse{
            Order:       withdrawal.OrderNumber,
            Sum:         withdrawal.Sum,
            ProcessedAt: withdrawal.ProcessedAt.Time,
        })
    }

    // Возвращаем ответ
    ctx.JSON(http.StatusOK, response) // 200
}
