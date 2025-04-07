package db

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestBalanceOperations(t *testing.T) {
	// Создаем тестового пользователя и аккаунт
	user := createRandomUser(t)
	accountID := createRandomLoyaltyAccount(t, user.ID)

	// Инициализируем начальный баланс (1000.00)
	initBalance := pgtype.Numeric{
		Int:   new(big.Int).SetInt64(100000),
		Exp:   -2,
		Valid: true,
	}
	// Инициализируем withdrawn (50.00)
	initWithdrawn := pgtype.Numeric{
		Int:   new(big.Int).SetInt64(5000),
		Exp:   -2,
		Valid: true,
	}

	_, err := testQueries.db.Exec(context.Background(),
		"UPDATE loyalty_accounts SET current_balance = $1, withdrawn_balance = $2 WHERE id = $3",
		initBalance, initWithdrawn, accountID)
	require.NoError(t, err)

	t.Run("GetBalanceByUserID", func(t *testing.T) {
		balance, err := testQueries.GetBalanceByUserID(context.Background(), pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		
		require.True(t, balance.CurrentBalance.Valid)
		require.Equal(t, initBalance.Int, balance.CurrentBalance.Int)
		
		require.True(t, balance.WithdrawnBalance.Valid)
		require.Equal(t, initWithdrawn.Int, balance.WithdrawnBalance.Int)
	})

	t.Run("UpdateBalance", func(t *testing.T) {
		// 1500.00
		newCurrent := pgtype.Numeric{
			Int:   new(big.Int).SetInt64(150000),
			Exp:   -2,
			Valid: true,
		}
		// 75.00
		newWithdrawn := pgtype.Numeric{
			Int:   new(big.Int).SetInt64(7500),
			Exp:   -2,
			Valid: true,
		}

		arg := UpdateBalanceParams{
			UserID:           pgtype.Int4{Int32: int32(user.ID), Valid: true},
			CurrentBalance:   newCurrent,
			WithdrawnBalance: newWithdrawn,
		}

		err := testQueries.UpdateBalance(context.Background(), arg)
		require.NoError(t, err)

		// Проверяем обновление
		updated, err := testQueries.GetBalanceByUserID(context.Background(), pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		require.Equal(t, newCurrent.Int, updated.CurrentBalance.Int)
		require.Equal(t, newWithdrawn.Int, updated.WithdrawnBalance.Int)
	})

	t.Run("CreateAndGetWithdrawals", func(t *testing.T) {
		// Создаем несколько списаний
		withdrawals := []struct {
			orderNumber string
			sum         float64
		}{
			{"ORDER123456", 100.50},
			{"ORDER654321", 200.75},
			{"ORDER111222", 50.25},
		}

		for _, w := range withdrawals {
			arg := CreateWithdrawalParams{
				UserID:      pgtype.Int4{Int32: int32(user.ID), Valid: true},
				OrderNumber: w.orderNumber,
				Sum:         w.sum,
			}
			err := testQueries.CreateWithdrawal(context.Background(), arg)
			require.NoError(t, err)
			time.Sleep(time.Millisecond) // Для разных временных меток
		}

		// Получаем список списаний
		result, err := testQueries.GetWithdrawalsByUserID(context.Background(), pgtype.Int4{Int32: int32(user.ID), Valid: true})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result), len(withdrawals))

		// Проверяем сортировку по времени (DESC)
		for i := 0; i < len(result)-1; i++ {
			require.True(t, result[i].ProcessedAt.Time.After(result[i+1].ProcessedAt.Time) || 
				result[i].ProcessedAt.Time.Equal(result[i+1].ProcessedAt.Time))
		}
	})

	t.Run("GetBalanceForNonExistentUser", func(t *testing.T) {
		_, err := testQueries.GetBalanceByUserID(context.Background(), pgtype.Int4{Int32: 999999, Valid: true})
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetWithdrawalsForUserWithoutWithdrawals", func(t *testing.T) {
		newUser := createRandomUser(t)
		result, err := testQueries.GetWithdrawalsByUserID(context.Background(), pgtype.Int4{Int32: int32(newUser.ID), Valid: true})
		require.NoError(t, err)
		require.Empty(t, result)
	})
}

// Вспомогательная функция для создания тестового лояльного аккаунта
func createRandomLoyaltyAccount(t *testing.T, userID int32) int32 {
	var accountID int32
	err := testQueries.db.QueryRow(context.Background(),
		"INSERT INTO loyalty_accounts (user_id, current_balance, withdrawn_balance) VALUES ($1, 0, 0) RETURNING id",
		userID).Scan(&accountID)
	require.NoError(t, err)
	return accountID
}