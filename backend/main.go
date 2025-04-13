package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"RandomGenerator/events"

	"github.com/gin-gonic/gin"
)

type Server struct {
	contractAddress string
}

// 测试用例：
//
//	curl -X POST http://localhost:8080/setContractAddress -H "Content-Type: application/json" \
//	     -d '{"address": "0x5FbDB2315678afecb367f032d93F642f64180aa3","timeout":10}'
func (s *Server) setContractAddress(c *gin.Context) {
	var json struct {
		Address string `json:"address" binding:"required"`
		Timeout int    `json:"timeout"` // 增加超时参数
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("setContractAddress : %v", json)

	s.contractAddress = json.Address

	event, err := events.NewEventListener(s.contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tm := json.Timeout
	if tm == 0 {
		tm = 1 * 60 // 默认超时时间为5分钟
	} else if json.Timeout < 0 {
		tm = 0 // 如果超时时间小于0，则不设置超时
	}
	log.Printf("setContractAddress timeout: %d", tm)
	go event.Listen(time.Duration(tm)) // 使用超时参数

	c.JSON(http.StatusOK, gin.H{"status": "contract address set"})
}

type Version struct {
	Ver string `json:"version"`
}

// 测试用例：
// curl -X GET http://localhost:58008/version
func (v *Version) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": v.Ver})
}

func main() {
	var port string
	// 增加参数设置，使用 -p 设置端口号
	flag.StringVar(&port, "p", "8888", "port")
	flag.Parse()

	ver := &Version{Ver: "1.0.0"}
	server := &Server{}

	log.Printf("Random Generator backend server started on port %s", port)
	r := gin.Default()
	r.POST("/setContractAddress", server.setContractAddress)
	r.GET("/version", ver.getVersion)
	r.Run(":" + port)
}
