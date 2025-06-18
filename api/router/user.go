package router

import (
	"github.com/Ian-zy0329/go-mall/api/controller"
	"github.com/Ian-zy0329/go-mall/common/middleware"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/user")
	g.GET("token/refresh", controller.TestRefreshToken)
	g.POST("register", controller.RegisterUser)
	g.POST("login", controller.LoginUser)
	g.DELETE("logout", middleware.AuthUser(), controller.LogoutUser)
}
