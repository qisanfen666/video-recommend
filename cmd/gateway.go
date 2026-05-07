package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	GoServerPort   = "8081"
	JavaServerPort = "8082"
)

func main() {
	r := gin.Default()

	//go服务
	r.GET("/api/videos/:id", forwardToGo)
	r.GET("/api/users/:id", forwardToGo)
	r.GET("/api/users/:id/history", forwardToGo)
	r.GET("/api/hot", forwardToGo)

	r.POST("/api/recommend", forwardToGo)
	r.POST("/api/similar-users", forwardToGo)

	log.Println("网关启动，监听 :8079")
	r.Run(":8079")
}

func forwardToGo(c *gin.Context) {
	url := "http://localhost:" + GoServerPort + c.Request.URL.Path

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
		return
	}

	req, err := http.NewRequest(c.Request.Method, url, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "构造转发请求失败"})
		return
	}
	if ct := c.GetHeader("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "网络错误",
		})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// func forwardToJava(c *gin.Context) {
// 	url := "http://localhost:" + JavaServerPort + c.Request.URL.Path

// 	body, _ := io.ReadAll(c.Request.Body)

// 	resp.err := http.Post(url, "application/json", bytes.NewBuffer(body))
// 	if err != nil {
// 		c.JSON(http.StatusServiceUnavailable, gin.H{
// 			"error": "网络不可用",
// 		})
// 		return
// 	}
// 	defer resp.Body.Close()

// 	respBody, _ := io.ReadAll(resp.Body)
// 	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
// }
