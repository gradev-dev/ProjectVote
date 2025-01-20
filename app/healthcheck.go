package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type DependencyHealthcheckOutput struct {
	Links  []string `json:"links"`
	Status string   `json:"status"`
}

type liveness struct {
	Version string   `json:"version"`
	Status  string   `json:"status"`
	Output  []string `json:"output"`
}

type statusHealthCheck struct {
	Version string   `json:"version"`
	Status  string   `json:"status"`
	Output  struct{} `json:"output,omitempty"`
}

type readiness struct {
	Version string                                 `json:"version"`
	Status  string                                 `json:"status"`
	Output  map[string]DependencyHealthcheckOutput `json:"output"`
}

func GetLiveness(c *gin.Context) {
	livenessResponse := liveness{
		Status:  "pass",
		Version: GetVersion(),
		Output:  make([]string, 0),
	}
	c.JSON(http.StatusOK, livenessResponse)
}

func GetStatus(c *gin.Context) {
	statusResponse := statusHealthCheck{
		Version: GetVersion(),
		Status:  "pass",
	}

	c.JSON(http.StatusOK, statusResponse)
}

func GetReadiness(c *gin.Context) {
	readinessResponse := readiness{
		Version: GetVersion(),
		Status:  "pass",
	}

	c.JSON(http.StatusOK, readinessResponse)
}
