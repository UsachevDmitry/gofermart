
package db

import (
	"context"
	"math/rand"
	"testing"
	"github.com/stretchr/testify/require"
)


func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Login)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Login, user2.Login)
	require.Equal(t, user1.Password, user2.Password)
}

// Вспомогательная функция для генерации случайных строк
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Login:    generateRandomString(10),
		Password: generateRandomString(12),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Login, user.Login)
	require.Equal(t, arg.Password, user.Password)
	require.NotZero(t, user.ID)

	return user
}
