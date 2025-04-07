package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestUpdateOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateLoyaltyTransaction", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)
		accountID := createTestAccount(t, q, user.ID)
		order := createTestOrder(t, q, user.ID, "PROCESSED")

		arg := CreateLoyaltyTransactionParams{
			AccountID:       pgtype.Int4{Int32: accountID, Valid: true},
			OrderID:         pgtype.Int4{Int32: order.ID, Valid: true},
			Points:          100.5,
			TransactionType: "ACCRUAL",
		}

		err = q.CreateLoyaltyTransaction(ctx, arg)
		require.NoError(t, err)
	})

	t.Run("GetLoyaltyAccountID", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)
		accountID := createTestAccount(t, q, user.ID)

		result, err := q.GetLoyaltyAccountID(ctx, pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		require.Equal(t, accountID, result)
	})

	t.Run("GetOrderID", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)
		order := createTestOrder(t, q, user.ID, "NEW")

		result, err := q.GetOrderID(ctx, order.OrderNumber)
		require.NoError(t, err)
		require.Equal(t, order.ID, result)
	})

	t.Run("GetOrderOwner", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)
		order := createTestOrder(t, q, user.ID, "NEW")

		result, err := q.GetOrderOwner(ctx, order.OrderNumber)
		require.NoError(t, err)
		require.Equal(t, user.ID, result.Int32)
	})

	t.Run("GetProcessedOrdersByUserID", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)

		// Очищаем возможные предыдущие заказы
		_, err = tx.Exec(ctx, "DELETE FROM orders WHERE user_id = $1", user.ID)
		require.NoError(t, err)

		// Создаем ровно 3 обработанных заказа
		for i := 0; i < 3; i++ {
			arg := SaveOrderParams{
				UserID:      pgtype.Int4{Int32: int32(user.ID), Valid: true},
				OrderNumber: generateRandomString(10),
				Status:      "PROCESSED",
				Accrual:     pgtype.Float8{Float64: float64((i+1)*100), Valid: true},
			}
			err := q.SaveOrder(ctx, arg)
			require.NoError(t, err)
		}

		orders, err := q.GetProcessedOrdersByUserID(ctx, pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		require.Len(t, orders, 3, "Должно быть ровно 3 обработанных заказа")
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		tx, err := testDB.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		q := New(tx)
		user := createRandomUser(t)
		order := createTestOrder(t, q, user.ID, "NEW")

		arg := UpdateOrderStatusParams{
			Status:      "PROCESSED",
			Accrual:     pgtype.Float8{Float64: 500.5, Valid: true},
			OrderNumber: order.OrderNumber,
		}

		err = q.UpdateOrderStatus(ctx, arg)
		require.NoError(t, err)

		// Проверяем обновление
		updatedOrder, err := q.GetOrderByNumber(ctx, order.OrderNumber)
		require.NoError(t, err)
		require.Equal(t, "PROCESSED", updatedOrder.Status)
	})
}

func createTestOrder(t *testing.T, q *Queries, userID int32, status string) Order {
	arg := CreateOrderParams{
		UserID:      pgtype.Int4{Int32: userID, Valid: true},
		OrderNumber: generateRandomString(10),
		Status:      status,
	}

	order, err := q.CreateOrder(context.Background(), arg)
	require.NoError(t, err)
	return order
}

func createTestAccount(t *testing.T, q *Queries, userID int32) int32 {
	var accountID int32
	err := q.db.QueryRow(context.Background(),
		"INSERT INTO loyalty_accounts (user_id, current_balance, withdrawn_balance) VALUES ($1, 0, 0) RETURNING id",
		userID).Scan(&accountID)
	require.NoError(t, err)
	return accountID
}