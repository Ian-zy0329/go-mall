package main

import (
	"github.com/Ian-zy0329/go-mall/api/router"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/config"
	"github.com/gin-gonic/gin"
)

func main() {
	if config.App.Env == enum.ModeProd {
		gin.SetMode(gin.ReleaseMode)
	}
	g := gin.New()
	router.RegisterRoutes(g)
	g.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
