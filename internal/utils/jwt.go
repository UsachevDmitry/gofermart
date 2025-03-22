package utils

import (
    "fmt"
    "errors"
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

// ValidateJWT проверяет JWT-токен и возвращает логин пользователя.
func ValidateJWT(tokenString string) (string, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("ошибка при парсинге токена: %v", err)
	}

	// Проверяем, что токен валиден
	if !token.Valid {
		return "", errors.New("неверный токен")
	}

	// Извлекаем claims (данные из токена)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("неверный формат claims")
	}

	// Извлекаем логин
	login, ok := claims["login"].(string)
	if !ok {
		return "", errors.New("логин не найден в токене")
	}

	return login, nil
}