package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestOrderCRUD(t *testing.T) {
	// Создаем пользователя для тестов
	user := createRandomUser(t)
	
	// Очищаем заказы перед тестами (если есть)
	_, err := testQueries.db.Exec(context.Background(), "DELETE FROM orders WHERE user_id = $1", user.ID)
	require.NoError(t, err)

	// Тест CreateOrder
	t.Run("CreateOrder", func(t *testing.T) {
		arg := CreateOrderParams{
			UserID:      pgtype.Int4{Int32: int32(user.ID), Valid: true},
			OrderNumber: generateRandomString(10),
			Status:      "NEW",
		}

		order, err := testQueries.CreateOrder(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, order)

		require.Equal(t, arg.UserID, order.UserID)
		require.Equal(t, arg.OrderNumber, order.OrderNumber)
		require.Equal(t, arg.Status, order.Status)
		require.False(t, order.UploadedAt.Time.IsZero())
	})

	// Тест GetOrderByNumber
	t.Run("GetOrderByNumber", func(t *testing.T) {
		// Сначала создаем заказ
		createArg := CreateOrderParams{
			UserID:      pgtype.Int4{Int32: int32(user.ID), Valid: true},
			OrderNumber: generateRandomString(10),
			Status:      "PROCESSING",
		}

		createdOrder, err := testQueries.CreateOrder(context.Background(), createArg)
		require.NoError(t, err)

		// Получаем заказ по номеру
		foundOrder, err := testQueries.GetOrderByNumber(context.Background(), createdOrder.OrderNumber)
		require.NoError(t, err)
		require.NotEmpty(t, foundOrder)

		require.Equal(t, createdOrder.ID, foundOrder.ID)
		require.Equal(t, createdOrder.OrderNumber, foundOrder.OrderNumber)
		require.Equal(t, createdOrder.Status, foundOrder.Status)
	})

	// Тест GetOrdersByUserID
	t.Run("GetOrdersByUserID", func(t *testing.T) {
		// Очищаем заказы перед тестом
		_, err := testQueries.db.Exec(context.Background(), "DELETE FROM orders WHERE user_id = $1", user.ID)
		require.NoError(t, err)

		// Создаем ровно 5 заказов для пользователя
		var orderNumbers []string
		for i := 0; i < 5; i++ {
			orderNumber := generateRandomString(10)
			arg := CreateOrderParams{
				UserID:      pgtype.Int4{Int32: int32(user.ID), Valid: true},
				OrderNumber: orderNumber,
				Status:      "NEW",
			}
			_, err := testQueries.CreateOrder(context.Background(), arg)
			require.NoError(t, err)
			orderNumbers = append(orderNumbers, orderNumber)
			time.Sleep(time.Millisecond * 10) // Для разных временных меток
		}

		// Получаем заказы пользователя
		orders, err := testQueries.GetOrdersByUserID(context.Background(), pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		require.Len(t, orders, 5, "Должно быть ровно 5 заказов")

		// Проверяем сортировку по времени (DESC)
		for i := 0; i < len(orders)-1; i++ {
			require.True(t, orders[i].UploadedAt.Time.After(orders[i+1].UploadedAt.Time) || 
				orders[i].UploadedAt.Time.Equal(orders[i+1].UploadedAt.Time))
		}

		// Проверяем структуру возвращаемых данных
		for _, order := range orders {
			require.NotEmpty(t, order.OrderNumber)
			require.Contains(t, orderNumbers, order.OrderNumber)
			require.NotEmpty(t, order.Status)
			require.False(t, order.UploadedAt.Time.IsZero())
		}
	})

	// Тест на несуществующий заказ
	t.Run("NonExistentOrder", func(t *testing.T) {
		_, err := testQueries.GetOrderByNumber(context.Background(), "nonexistent123")
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}