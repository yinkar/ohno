package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type input struct {
	Url	string `json:"url"`
}

type output struct {
	ScanId string `json:"scan_id"`
}

func main() {
	router := gin.Default()

	router.GET("ping", ping)
	router.POST("newscan", newScan)

	router.Run("localhost:8080")

	fmt.Println("Test")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func newScan(c *gin.Context) {
	var newInput input
	var newOutput output

	if err := c.BindJSON(&newInput); err != nil {
		return
	}

	scanId := uuid.New()

	newOutput.ScanId = scanId.String()

	c.IndentedJSON(http.StatusCreated, newOutput)
}