package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("ping", ping)

	router.Run("localhost:8080")

	fmt.Println("Test")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}
