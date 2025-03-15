package api

import (
	db "db/sqlc"
	"utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	//"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Login        string `json:"login"`
	Password string `json:"password"`
}

type createUserResponce struct {
	Login             string `json:"login"`
	Password  	  string `json:"password"`
}

func (server *Server) CreateUser(ctx *gin.Context) {
	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponce(err))
		return
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponce(err))
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
			switch pgErr.Message {
			case "23505":
				ctx.JSON(http.StatusForbidden, errorResponce(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponce(err))
		return
	}
	rsp := createUserResponce{
		Login:          user.Login,
		Password: 	user.Password,
	}
	ctx.JSON(http.StatusOK, rsp)
}
