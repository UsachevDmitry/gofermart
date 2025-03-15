package api

import (
	db "db/sqlc"
	"errors"
	"net/http"
	"utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	//"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
}

// type createUserResponce struct {
// 	Login             string `json:"login"`
// 	Password  	  string `json:"password"`
// }

func (server *Server) CreateUser(ctx *gin.Context) {
	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	arg := db.CreateUserParams{
		Login: 				req.Login,
		Password:       hashedPassword,
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				ctx.Status(http.StatusConflict)
				return
			}
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}
	// Генерация JWT-токена
	token, err := utils.GenerateJWT(user.Login)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	// Успешный ответ с токеном
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
	//ctx.Status(http.StatusOK)
}
