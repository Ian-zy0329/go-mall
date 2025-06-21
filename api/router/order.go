package router

import (
	"github.com/Ian-zy0329/go-mall/api/controller"
	"github.com/Ian-zy0329/go-mall/common/middleware"
	"github.com/gin-gonic/gin"
)

func registerOrderRouter(rg *gin.RouterGroup) {
	g := rg.Group("/order")
	g.Use(middleware.AuthUser())
	g.POST("create", controller.OrderCreate)
	g.GET("user-order", controller.UserOrders)
	g.GET(":order_no/info", controller.OrderInfo)
	g.PATCH(":order_no/cancel", controller.CancelOrder)
}
