package order

import (
	"iswift-go-project/internal/middleware"
	dto "iswift-go-project/internal/order/dto"
	usecase "iswift-go-project/internal/order/usecase"
	"iswift-go-project/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	usecase usecase.OrderUseCase
}

func NewOrderHandler(usecase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{usecase}
}

func (handler *OrderHandler) Route(r *gin.RouterGroup) {
	orderHandler := r.Group("/api/v1")

	orderHandler.Use(middleware.AuthJwt)
	{
		orderHandler.POST("/orders", handler.Create)
		orderHandler.GET("/orders", handler.FindAllByUserId)
	}
}

func (handler *OrderHandler) Create(ctx *gin.Context) {
	var input dto.OrderRequestBody

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Response(http.StatusBadRequest, "bad request", err.Error()))
		ctx.Abort()
		return
	}

	user := utils.GetCurrentUser(ctx)

	input.UserID = user.ID
	input.Email = user.Email

	data, err := handler.usecase.Create(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Response(http.StatusInternalServerError, "internal server error", err.Error()))
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, utils.Response(http.StatusOK, "ok", data))
}

func (handler *OrderHandler) FindAllByUserId(ctx *gin.Context) {
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	limit, _ := strconv.Atoi(ctx.Param("limit"))

	user := utils.GetCurrentUser(ctx)
	data := handler.usecase.FindAllByUserId(offset, limit, int(user.ID))

	ctx.JSON(http.StatusOK, utils.Response(http.StatusOK, "ok", data))
}
