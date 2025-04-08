package api

import (
	db "db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"net/http"
	"time"
	"utils"
)

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

	// Сохраняем заказ со статусом NEW/REGISTERED
	if err := server.store.SaveOrder(ctx, db.SaveOrderParams{
		OrderNumber: string(orderNumber),
		UserID:      userID,
		Status:      "REGISTERED",
		Accrual:     pgtype.Float8{Valid: false},
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить заказ"})
		return
	}

	// Запускаем фоновую обработку
	go func() {
		if err := server.orderService.PollOrderStatus(string(orderNumber)); err != nil {
			log.Printf("Failed to process order %s: %v\n", string(orderNumber), err)
		}
	}()

	ctx.JSON(http.StatusAccepted, gin.H{"message": "заказ принят в обработку"})
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
