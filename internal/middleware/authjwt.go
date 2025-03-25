package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "utils"
)

//var jwtSecret = []byte("secret-key")

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Получаем токен из cookie
		token, err := ctx.Cookie("token")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "токен отсутствует"})
			ctx.Abort()
			return
		}

		// Проверяем токен
		login, err := utils.ValidateJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
			ctx.Abort()
			return
		}

		// Добавляем логин в контекст
		ctx.Set("login", login)
		ctx.Next()
	}
}