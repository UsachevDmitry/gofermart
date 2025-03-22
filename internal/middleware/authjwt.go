package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("secret-key") // ваш секретный ключ

// AuthMiddleware проверяет JWT-токен
func AuthMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        authHeader := ctx.GetHeader("Authorization")
        if authHeader == "" {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "токен отсутствует"})
            ctx.Abort()
            return
        }

        // Извлечение токена из заголовка
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
            ctx.Abort()
            return
        }

        // Парсинг и проверка токена
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
            ctx.Abort()
            return
        }

        // Добавление данных из токена в контекст
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            ctx.Set("login", claims["login"])
        }

        ctx.Next()
    }
}