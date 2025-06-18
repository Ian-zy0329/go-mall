package router

import (
	"github.com/Ian-zy0329/go-mall/api/controller"
	"github.com/Ian-zy0329/go-mall/common/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	engine.Use(middleware.StartTrace(), middleware.LogAccess(), middleware.GinPanicRecovery())
	routeGroup := engine.Group("")
	registerBuildingRoutes(routeGroup)
	registerUserRoutes(routeGroup)
}

func registerBuildingRoutes(routeGroup *gin.RouterGroup) {
	g := routeGroup.Group("/building")
	g.GET("ping", controller.TestPing)
	g.GET("config-read", controller.TestConfigRead)
	g.POST("logger-test", controller.TestLogger)
	g.POST("access-log-test", controller.TestAccessLog)
	g.GET("panic-log-test", controller.TestPanicLog)
	g.GET("customized-error-test", controller.TestAppError)
	g.GET("response-obj", controller.TestResponseObj)
	g.GET("response-list", controller.TestResponseList)
	g.GET("response-err", controller.TestResponseError)
	g.GET("gorm-logger-test", controller.TestGormLogger)
	g.POST("create-demo-order", controller.TestCreateDemoOrder)
	g.GET("httptool-get-test", controller.TestForHttpToolGet)
	g.GET("httptool-post-test", controller.TestForHttpToolPost)
	g.GET("token-make-test", controller.TestMakeToken)
	g.GET("token-auth-test", middleware.AuthUser(), controller.TestAuthToken)
	g.GET("token-refresh-test", controller.TestRefreshToken)
}
