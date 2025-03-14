package main

import (
	"net/http"

	"RandomGenerator/events"

	"github.com/gin-gonic/gin"
)

type Server struct {
	contractAddress string
}

// 测试用例：
//
//	curl -X POST http://localhost:8080/setContractAddress -H "Content-Type: application/json" \
//	     -d '{"address": "0x5FbDB2315678afecb367f032d93F642f64180aa3"}'
func (s *Server) setContractAddress(c *gin.Context) {
	var json struct {
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.contractAddress = json.Address

	events, err := events.NewEventListener(s.contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go events.Listen()

	c.JSON(http.StatusOK, gin.H{"status": "contract address set"})
}

type Version struct {
	Ver string `json:"version"`
}

// 测试用例：
// curl -X GET http://localhost:8080/version
func (v *Version) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": v.Ver})
}

func main() {
	ver := &Version{Ver: "1.0.0"}
	server := &Server{}

	r := gin.Default()
	r.POST("/setContractAddress", server.setContractAddress)
	r.GET("/version", ver.getVersion)
	r.Run(":8080")
}
