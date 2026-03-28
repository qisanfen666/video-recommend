package main

import (
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
	r.GET("/api/video/:id", forwardToGo)
	r.GET("/api/users/:id", forwardToGo)
	r.GET("/api/users/:id/history", forwardToGo)
	r.GET("/api/hot", forwardToGo)

	//java服务
	r.GET("/api/recommend", forwardToGo)
	r.GET("/api/similar-users", forwardToGo)

	log.Println("网关启动，监听 :8080")
	r.Run(":8080")
}

func forwardToGo(c *gin.Context) {
	url := "http://localhost:" + GoServerPort + c.Request.URL.Path

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "网络错误",
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
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
