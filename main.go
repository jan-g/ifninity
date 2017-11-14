package main

import "github.com/gin-gonic/gin"

import f "github.com/jan-g/ifninity/function"

func main() {
	r := gin.Default()
	r.GET("/ping", ping)
	r.POST("/ping", ping)
	r.GET("/r/t/fast-function", f.Fast)
	r.POST("/r/t/fast-function", f.Fast)
	r.GET("/r/t/slow-function", f.Slow)
	r.POST("/r/t/slow-function", f.Slow)
	r.GET("/r/t/flow-load-test-vista", f.Vista)
	r.POST("/r/t/flow-load-test-vista", f.Vista)
	r.Run() // listen and serve on 0.0.0.0:8080
}


func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
