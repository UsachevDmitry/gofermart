package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"utils"
)

type loginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"}) // 400
		return
	}

	// Получаем пользователя из базы данных
	user, err := server.store.GetUser(ctx, req.Login)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверная пара логин/пароль"}) // 401
		return
	}

	// Проверяем пароль
	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверная пара логин/пароль"}) // 401
		return
	}

	// Генерация JWT-токена
	token, err := utils.GenerateJWT(user.Login)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при генерации токена"}) // 500
		return
	}

	// Устанавливаем токен в cookies
	ctx.SetCookie("token", token, 3600, "/", "", false, true)

	// Успешный ответ
	ctx.JSON(http.StatusOK, gin.H{"token": token}) // 200
}
