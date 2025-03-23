package api

import (
    //"errors"
    "log"
    "net/http"
    "utils"
	"github.com/jackc/pgx/v5/pgtype"
	//"github.com/jackc/pgx/v5/pgconn"
    "github.com/gin-gonic/gin"
    db "db/sqlc"
	"time"
)

// func (server *Server) uploadOrder(ctx *gin.Context) {
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
//     // Чтение номера заказа из тела запроса
//     orderNumber, err := ctx.GetRawData()
//     if err != nil {
//         ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
//         return
//     }

//     // Проверка номера заказа с помощью алгоритма Луна
//     if !utils.IsValidLuhn(string(orderNumber)) {
//         ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "неверный формат номера заказа"}) // 422
//         return
//     }

//     // Сохранение номера заказа в базе данных
//     _ , err = server.store.CreateOrder(ctx, db.CreateOrderParams{
//         UserID:      userIDInt4,
//         OrderNumber: string(orderNumber),
//         Status:      "NEW", // Статус по умолчанию
//     })
//     if err != nil {
//         var pgErr *pgconn.PgError
//         if errors.As(err, &pgErr) {
//             switch pgErr.Code {
//             case "23505": // Ошибка уникальности (номер заказа уже существует)
//                 // Проверка, был ли номер заказа загружен этим пользователем
//                 existingOrder, err := server.store.GetOrderByNumber(ctx, string(orderNumber))
//                 if err != nil {
//                     ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//                     return
//                 }
//                 if existingOrder.UserID.Int32 == userID {
//                     ctx.JSON(http.StatusOK, gin.H{"message": "номер заказа уже был загружен этим пользователем"}) // 200
//                 } else {
//                     ctx.JSON(http.StatusConflict, gin.H{"error": "номер заказа уже был загружен другим пользователем"}) // 409
//                 }
//                 return
//             }
//         }
//         ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
//         return
//     }

//     // Успешный ответ
//     ctx.JSON(http.StatusAccepted, gin.H{"message": "новый номер заказа принят в обработку"}) // 202
// }
func (server *Server) uploadOrder(ctx *gin.Context) {
    // Проверка аутентификации
    login, exists := ctx.Get("login")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"}) // 401
        return
    }

    // Получаем пользователя из базы данных
    user, err := server.store.GetUser(ctx, login.(string))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить данные пользователя"}) // 500
        return
    }
    userID := pgtype.Int4{Int32: user.ID, Valid: true}

    // Чтение номера заказа из тела запроса
    orderNumber, err := ctx.GetRawData()
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
        return
    }

    // Проверка номера заказа с помощью алгоритма Луна
    if !utils.IsValidLuhn(string(orderNumber)) {
        ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "неверный формат номера заказа"}) // 422
        return
    }

    // Проверка, был ли заказ уже загружен
    existingOrder, err := server.store.GetOrderByNumber(ctx, string(orderNumber))
    if err == nil {
        if existingOrder.UserID == userID {
            ctx.JSON(http.StatusOK, gin.H{"message": "заказ уже был загружен этим пользователем"}) // 200
            return
        } else {
            ctx.JSON(http.StatusConflict, gin.H{"error": "заказ уже был загружен другим пользователем"}) // 409
            return
        }
    }

    // Сохранение заказа в базу данных
    if err := server.store.SaveOrder(ctx, db.SaveOrderParams{
        OrderNumber: string(orderNumber),
        UserID:      userID,
        Status:      "NEW",
        Accrual:     pgtype.Float8{Float64: 0, Valid: true},
    }); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить заказ"}) // 500
        return
    }

    // Запуск фоновой задачи для опроса статуса заказа
    go func() {
        if err := server.orderService.PollOrderStatus(string(orderNumber)); err != nil {
            // Логируем ошибку
            log.Printf("Failed to poll order status: %v\n", err)
        }
    }()

    ctx.JSON(http.StatusAccepted, gin.H{"message": "заказ принят в обработку"}) // 202
}

type OrderResponse struct {
    Number     string    `json:"number"`
    Status     string    `json:"status"`
    Accrual    float64   `json:"accrual,omitempty"`
    UploadedAt time.Time `json:"uploaded_at"`
}

func (server *Server) getOrders(ctx *gin.Context) {
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

    // Получаем список заказов пользователя
    orders, err := server.store.GetOrdersByUserID(ctx.Request.Context(), userIDInt4)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
        return
    }

    // Если заказов нет, возвращаем 204
    if len(orders) == 0 {
        ctx.JSON(http.StatusNoContent, "") // 204
        return
    }

    // Форматируем ответ
    var response []OrderResponse
    for _, order := range orders {
        response = append(response, OrderResponse{
            Number:     order.OrderNumber,
            Status:     order.Status,
            Accrual:    order.Accrual.Float64,
            UploadedAt: order.UploadedAt.Time,
        })
    }

    // Возвращаем ответ
    ctx.JSON(http.StatusOK, response) // 200
}