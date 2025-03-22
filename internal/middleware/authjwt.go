package middleware

import (
    "net/http"
    //"strings"
    "github.com/gin-gonic/gin"
    //"github.com/golang-jwt/jwt/v5"
    "utils"
)

var jwtSecret = []byte("secret-key") // ваш секретный ключ

// // AuthMiddleware проверяет JWT-токен
// func AuthMiddleware() gin.HandlerFunc {
//     return func(ctx *gin.Context) {
//         authHeader := ctx.GetHeader("Authorization")
//         if authHeader == "" {
//             ctx.JSON(http.StatusUnauthorized, gin.H{"error": "токен отсутствует"})
//             ctx.Abort()
//             return
//         }

//         // Извлечение токена из заголовка
//         tokenString := strings.TrimPrefix(authHeader, "Bearer ")
//         if tokenString == authHeader {
//             ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
//             ctx.Abort()
//             return
//         }

//         // Парсинг и проверка токена
//         token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//             return jwtSecret, nil
//         })
//         if err != nil || !token.Valid {
//             ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
//             ctx.Abort()
//             return
//         }

//         // Добавление данных из токена в контекст
//         if claims, ok := token.Claims.(jwt.MapClaims); ok {
//             ctx.Set("login", claims["login"])
//         }

//         // Получаем токен из cookie
// 		token_cookie, err_cookie := ctx.Cookie("token")
// 		if err_cookie != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "токен отсутствует"})
// 			ctx.Abort()
// 			return
// 		}

// 		// Проверяем токен
// 		login, err := utils.ValidateJWT(token_cookie)
// 		if err != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
// 			ctx.Abort()
// 			return
// 		}

// 		// Добавляем логин в контекст
// 		ctx.Set("login", login)

//         ctx.Next()
//     }
// }

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