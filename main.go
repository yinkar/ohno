package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-git/go-git/v5"
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

	// Get URL
	if err := c.BindJSON(&newInput); err != nil {
		log.Fatal(err)
	}

	scanId := uuid.New()

	newOutput.ScanId = scanId.String()

	// Clone repo to /tmp
	clonePath := filepath.Join("/tmp/src", newOutput.ScanId)

	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:	newInput.Url,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatal(err)
	}

	// Return Scan ID output
	c.IndentedJSON(http.StatusCreated, newOutput)
}