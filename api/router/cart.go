package router

import (
	"github.com/Ian-zy0329/go-mall/api/controller"
	"github.com/Ian-zy0329/go-mall/common/middleware"
	"github.com/gin-gonic/gin"
)

func registerCartRouter(rg *gin.RouterGroup) {
	g := rg.Group("/cart/")
	g.Use(middleware.AuthUser())
	g.POST("addCartItem", controller.AddCartItem)
	g.GET("/item/check-bill", controller.CheckCartItemBill)
	g.PATCH("update-item", controller.UpdateCartItem)
	g.GET("item", controller.UserCartItems)
	g.DELETE("item/:item_id", controller.DeleteCartItem)
}
