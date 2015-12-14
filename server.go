package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func startServer() {
	bindAddress := fmt.Sprintf("%s:%d", config.Address, config.Port)

	log.Printf("Starting server on %s\n", bindAddress)

	gin.SetMode(gin.ReleaseMode)

	router := setupRouter()

	go func(r *gin.Engine) {
		if err := r.Run(bindAddress); err != nil {
			log.Fatalf("Can't start server: %v", err)
		}
	}(router)
}

func setupRouter() (router *gin.Engine) {
	router = gin.New()

	router.Use(gin.Recovery())

	if config.Username != "" && config.Password != "" {
		router.Use(
			gin.BasicAuth(gin.Accounts{config.Username: config.Password}),
		)
	}

	router.POST("/:application/put", createLogHandler)
	router.GET("/:application/get", getLogsHandler)
	router.GET("/:application/stats", appStatsHandler)

	return
}
