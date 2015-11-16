package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func startServer() {
	bindAddress := config.Server.Address + ":" + config.Server.Port

	log.Printf("Starting server on %s\n", bindAddress)

	gin.SetMode(gin.ReleaseMode)

	router := setupRouter()

	if err := router.Run(bindAddress); err != nil {
		log.Fatalf("Can't start server: %v", err)
	}
}

func setupRouter() (router *gin.Engine) {
	router = gin.New()

	router.Use(
		gin.Recovery(),
		gin.BasicAuth(gin.Accounts{config.Auth.User: config.Auth.Password}),
	)

	router.POST("/:application/put", createLogHandler)
	router.GET("/:application/get", getLogsHandler)
	router.GET("/:application/stats", appStatsHandler)

	return
}
