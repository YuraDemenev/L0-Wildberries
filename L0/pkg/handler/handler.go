package handler

import "github.com/gin-gonic/gin"

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.LoadHTMLGlob("../templates/*")

	orderPage := router.Group("/order")
	{
		orderPage.GET("/GetInfo", h.getOrderInfo)
		orderPage.POST("/GetInfo", h.postOrderInfo)

		orderPage.GET("/CreateProduct", h.getCreateProduct)
		orderPage.POST("/CreateProduct", h.postCreateProduct)
	}
	return router
}
