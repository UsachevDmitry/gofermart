package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("secret-key") 

// GenerateJWT создает JWT-токен для пользователя
func GenerateJWT(login string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "login": login,
        "exp":   time.Now().Add(time.Hour * 24).Unix(), // Токен действителен 24 часа
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}