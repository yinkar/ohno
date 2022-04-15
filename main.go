package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"path/filepath"
	"context"
	"io"
	"io/ioutil"
	"database/sql"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-git/go-git/v5"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	_ "github.com/mattn/go-sqlite3"
)

type input struct {
	Url	string `json:"url"`
}

type output struct {
	ScanId string `json:"scan_id"`
}

type errorType struct {
	Error bool `json:"error"`
	Message string `json:"message"`
}

type scan struct {
	Id string `json:"id"`
	Time string `json:"time"`
	Content string `json:"content"`
}

func handleError(err error, msg string, c *gin.Context) {
	newError := errorType{Error: true, Message: msg}
	c.IndentedJSON(http.StatusCreated, newError)
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

func main() {
	router := gin.Default()

	router.GET("/ping", ping)
	router.POST("/newscan", createScan)
	router.GET("/scan/:scan_id", viewScan)

	router.Run("localhost:8080")

	fmt.Println("Test")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func createScan(c *gin.Context) {
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
		handleError(err, "Repo cloning error.", c)
		return
	}

	// Docker
	// Simulating "docker run --rm -v ${PWD}:/code opensorcery/bandit -r /code"
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		handleError(err, "Context error.", c)
		return
	}

	reader, err := cli.ImagePull(ctx, "opensorcery/bandit", types.ImagePullOptions{})
	if err != nil {
		handleError(err, "Bandit Docker image pull error.", c)
		return
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "opensorcery/bandit",
		Cmd: []string{"-r", "/code", "-f", "json"},
	}, &container.HostConfig{
			LogConfig: container.LogConfig{
				Type:   "json-file",
				Config: map[string]string{},
			},
			Binds: []string{
				fmt.Sprintf("%s:/code", clonePath),
			},
		}, nil, nil, "")
	if err != nil {
		handleError(err, "Bandit container create error.", c)
		return
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		handleError(err, "Bandit container start error.", c)
		return
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			handleError(err, "Bandit container wait error.", c)
			return
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		handleError(err, "Bandit container log error.", c)
		return
	}

	//	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	f, err := os.Create("/tmp/clogs")
	io.Copy(f, out)
	f.Close()

	data, _ := ioutil.ReadFile("/tmp/clogs")
	testResult := string(data)

	fmt.Println(testResult)


	// Save to database
	db := databaseConnection(c);

	stmt, err := db.Prepare("INSERT INTO scans(id, time, content) VALUES(?, ?, ?)")
	if err != nil {
		handleError(err, "Query error.", c)
		return
	}

	_, err = stmt.Exec(newOutput.ScanId, time.Now(), testResult)
	if err != nil {
		handleError(err, "Database insert error.", c)
		return
	}

	db.Close()

	// Return Scan ID output
	c.IndentedJSON(http.StatusCreated, newOutput)
}

func viewScan(c *gin.Context) {
	scanId := c.Param("scan_id")

	var currentScan scan

	db := databaseConnection(c);

	stmt, err := db.Prepare("SELECT id, time, content FROM scans WHERE id = ?")
	if err != nil {
		handleError(err, "Query error.", c)
		return
	}

	err = stmt.QueryRow(scanId).Scan(&currentScan.Id, &currentScan.Time, &currentScan.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(err, "Database select error.", c)
			return
		}

		c.JSON(http.StatusOK, currentScan.Content)
		return
	}

	stmt.Close()

	db.Close()

	c.JSON(http.StatusOK, currentScan.Content)
}