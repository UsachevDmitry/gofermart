package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"net/http"
	"time"
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

type serviceError string

func (e serviceError) Error() string { return string(e) }

const ErrInsufficientFunds = serviceError("insufficient funds")

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
	err := numeric.Scan(fmt.Sprintf("%f", f))
	return numeric, err
}

func (s *Server) getBalance(ctx *gin.Context) {
	// Проверка аутентификации
	login, exists := ctx.Get("login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"})
		return
	}

	// Получаем пользователя
	user, err := s.store.GetUser(ctx, login.(string))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		} else {
			log.Printf("Failed GetUser: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	// Получаем баланс через OrderService
	current, withdrawn, err := s.orderService.GetUserBalance(ctx.Request.Context(), user.ID)
	if err != nil {
		log.Printf("Failed GetUserBalance: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить баланс"})
		return
	}

	// Преобразуем float64 в pgtype.Numeric для ответа
	currentNumeric, err := float64ToNumeric(current)
	if err != nil {
		log.Printf("Failed currentNumeric %#v to float64ToNumeric: %v", current, err)

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	withdrawnNumeric, err := float64ToNumeric(withdrawn)
	if err != nil {
		log.Printf("Failed withdrawnNumeric to float64ToNumeric: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, BalanceResponse{
		Current:   currentNumeric,
		Withdrawn: withdrawnNumeric,
	})
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

	// Преобразуем сумму из pgtype.Numeric в float64
	sum, err := numericToFloat64(req.Sum)
	if err != nil {
		log.Printf("Failed to convert sum: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат суммы"}) // 400
		return
	}

	// // Выполняем списание через OrderService
	// if err := s.orderService.Withdraw(ctx.Request.Context(), user.ID, req.Order, sum); err != nil {
	// 	if err.Error() == "insufficient funds" {
	// 		ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "недостаточно средств"}) // 402
	// 	} else {
	// 		log.Printf("Failed to withdraw: %v", err)
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось выполнить списание"}) // 500
	// 	}
	// 	return
	// }
	
	// Выполняем списание через OrderService
	if err := s.orderService.Withdraw(ctx.Request.Context(), user.ID, req.Order, sum); err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			ctx.JSON(http.StatusPaymentRequired, gin.H{"error": "недостаточно средств"}) // 402
		} else {
			log.Printf("Failed to withdraw: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось выполнить списание"}) // 500
		}
		return
	}



	ctx.JSON(http.StatusOK, gin.H{"message": "списание выполнено успешно"}) // 200
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
		log.Printf("Failed to GetUser: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
		return
	}

	// Получаем список списаний через OrderService
	withdrawals, err := s.orderService.GetUserWithdrawals(ctx.Request.Context(), user.ID)
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
		sumNumeric, err := float64ToNumeric(withdrawal.Sum)
		if err != nil {
			log.Printf("Failed to convert withdrawal sum: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"}) // 500
			return
		}

		response = append(response, WithdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         sumNumeric,
			ProcessedAt: withdrawal.ProcessedAt.Time,
		})
	}

	// Возвращаем ответ
	ctx.JSON(http.StatusOK, response) // 200
}
