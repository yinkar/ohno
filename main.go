package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yinkar/ohno/src"
)

func main() {
	router := gin.Default()

	router.GET("/ping", src.Ping)
	router.POST("/newscan", src.CreateScan)
	router.GET("/scan/:scan_id", src.ViewScan)

	router.Run("localhost:8080")
}