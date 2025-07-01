package main

import (
	"context"
	"github.com/Ian-zy0329/go-mall/api/router"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if config.App.Env == enum.ModeProd {
		gin.SetMode(gin.ReleaseMode)
	}
	g := gin.New()
	router.RegisterRoutes(g)
	server := http.Server{
		Addr:    ":8080",
		Handler: g,
	}
	log := logger.New(context.Background())
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("ShutdownServerError", "err", err)
		}
	}()
	log.Info("Starting GO MALL HTTP server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Info("Server closed under request")
		} else {
			log.Error("Server closed unexpect")
		}
	}
}
