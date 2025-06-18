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
	g.POST("password/apply-reset", controller.PasswordResetApply)
	g.POST("password/reset", controller.PasswordReset)
	g.GET("info", middleware.AuthUser(), controller.UserInfo)
	g.PATCH("info", middleware.AuthUser(), controller.UpdateUserInfo)
	g.POST("address", middleware.AuthUser(), controller.AddUserAddress)
	g.GET("address", middleware.AuthUser(), controller.GetUserAddresses)
	g.PATCH("address/:address_id", middleware.AuthUser(), controller.UpdateUserAddress)
	g.GET("address/:address_id", middleware.AuthUser(), controller.GetSingleAddress)
	g.DELETE("address/:address_id", middleware.AuthUser(), controller.DeleteUserAddress)
}
