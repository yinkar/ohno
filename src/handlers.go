package src

import (
	"fmt"
	"net/http"
	"os"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func handleError(err error, msg string, c *gin.Context) {
	newError := ErrorType{Error: true, Message: msg}
	c.IndentedJSON(http.StatusBadRequest, newError)
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", err))
}

func databaseConnection(c *gin.Context) (*sql.DB) {
	db, err := sql.Open("sqlite3", "./ohno.db")
	if err != nil {
		handleError(err, "Database connection error.", c)
		return db
	}

	return db
}