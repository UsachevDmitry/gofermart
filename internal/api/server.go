package api

import (
	db "db/sqlc"
	"net/http"
	"middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	store *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Public routes (не требуют аутентификации)
	router.POST("/api/user/register", server.CreateUser)
	router.POST("/api/user/login", server.loginUser)

	// Protected routes (требуют аутентификации)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware()) // Добавляем middleware для всех маршрутов в группе /api
	{
		protected.GET("/user/get/:login", server.GetUser)
		protected.POST("/user/orders", server.uploadOrder)
		protected.GET("/user/orders", server.getOrders)
		protected.GET("/user/balance", server.getBalance)
		protected.POST("/user/balance/withdraw", server.withdrawBalance)
		protected.GET("/user/withdrawals", server.getWithdrawals)
	}
	server.router = router
	return server
}

// errorResponce return gin.H -> map[string]interface{}
func errorResponce(err error) gin.H {
	return gin.H{"error": err.Error()}
}

// Start server method
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

type getUserRequest struct {
	Login string `uri:"login" binding:"required"`
}

func (server *Server) GetUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponce(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Login)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponce(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}