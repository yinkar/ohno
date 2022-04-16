package src

import (
	"net/http"
	"os"
	"path/filepath"
	"context"
	"io"
	"time"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-git/go-git/v5"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/docker/docker/pkg/stdcopy"
)

/**
	Ping
	
	/ping
*/
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

/**
	Create a scan
	
	/newscan
*/
func CreateScan(c *gin.Context) {
	var newInput Input
	var newOutput Output

	// Get URL
	if err := c.BindJSON(&newInput); err != nil {
		handleError(err, "Input error.", c)
		return
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
	
    dst := &bytes.Buffer{}
	stdcopy.StdCopy(dst, os.Stderr, out)

	testResultRaw := dst.String()

	testResult := &Scan{}
	err = json.Unmarshal([]byte(testResultRaw), testResult)
	if err != nil {
		handleError(err, "Result parse error.", c)
		return
	}

	// Save to database
	db := databaseConnection(c)
	defer db.Close()

	for _, result := range testResult.Results {
		stmt, err := db.Prepare("INSERT INTO results(scan_id, scan_error, code, filename, issue_severity, created_at) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			handleError(err, "Query error.", c)
			return
		}

		_, err = stmt.Exec(newOutput.ScanId, "", result.Code, result.Filename, result.IssueSeverity, time.Now())
		if err != nil {
			handleError(err, "Database insert error.", c)
			return
		}
	}

	// Return Scan ID output
	c.IndentedJSON(http.StatusCreated, newOutput)
}

/**
	View a scan

	/scan/:scan_id
*/
func ViewScan(c *gin.Context) {
	scanId := c.Param("scan_id")

	var result Result
	var scan Scan
	
	db := databaseConnection(c);
	defer db.Close()

	scan.Safety = true
	highSevertyCount := 0

	rows, err := db.Query("SELECT scan_id, code, filename, issue_severity, created_at FROM results WHERE scan_id = ?", scanId)
	if err != nil {
		handleError(err, "Fetching scan results error.", c)
		return
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&result.ScanId, &result.Code, &result.Filename, &result.IssueSeverity, &result.CreatedAt)
		if err != nil {
			continue
		}

		scan.Results = append(scan.Results, result)

		if result.IssueSeverity == "HIGH" {
			highSevertyCount++
		}
	}
	err = rows.Err()
	if err != nil {
		handleError(err, "Row iterating error.", c)
		return
	}

	// Set safety negative by HIGH severity count
	if highSevertyCount > 1 {
		scan.Safety = false
	}

	c.JSON(http.StatusOK, scan)
}