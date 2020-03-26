package main

import (
	"log"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)


// Logger ...
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()

		latency := end.Sub(start)

		path := c.Request.URL.Path

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		glog.Infof("| %3d | %13v | %15s | %s  %s |",
			statusCode,
			latency,
			clientIP,
			method, path,
		)
	}

}

func init()  {
	log.SetPrefix("TRACE:")
}

func main()  {
	fmt.Println("ssss")
	log.Println("aaaaaaa")
}
