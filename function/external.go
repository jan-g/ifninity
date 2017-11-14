package function

import "io"
import "time"
import "github.com/gin-gonic/gin"


/*
public byte[] handleRequest(byte[] input) {
    try {
		Thread.sleep(100);
	}
	catch (Exception e) {
		// Ok
	}
	return input;
}
*/

func Fast(c *gin.Context) {
	time.Sleep(100 * time.Millisecond)
	c.Header("Content-Type", "application/octet-stream")
	c.Writer.WriteHeaderNow()
	io.Copy(c.Writer, c.Request.Body)
}


func Slow(c *gin.Context) {
	time.Sleep(5 * time.Second)
	c.Header("Content-Type", "application/octet-stream")
	c.Writer.WriteHeaderNow()
	io.Copy(c.Writer, c.Request.Body)
}