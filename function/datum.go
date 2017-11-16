package function

import (
	"github.com/gin-gonic/gin"
	"github.com/jan-g/ifninity/flow"
)


func returnDatum(c *gin.Context, datumType string, success bool, contentType string, body []byte, headers map[string]string) {
	c.Writer.WriteHeaderNow()
	s := "success"
	if !success {
		s = "failure"
	}
	ct := ""
	if contentType != "" {
		ct = "Content-Type: " + contentType + "\r\n"
	}
	hs := ""
	for k, v := range headers {
		hs += k + ": " + v + "\r\n"
	}
	c.Writer.Write([]byte(
		"HTTP/1.1 200\r\n" +
		"FnProject-DatumType: " + datumType + "\r\n" +
		"FnProject-ResultStatus: " + s + "\r\n" +
		ct +
		hs +
		"\r\n"))
	c.Writer.Write(body)
}


func returnEmpty(c *gin.Context) {
	returnDatum(c, "empty", true, "", nil, nil)
}

func returnStage(c *gin.Context, stage flow.Stage) {
	returnDatum(c, "stageref", true, "", nil, map[string]string{"FnProject-StageId": stage.Id()})
}

func returnBlob(c *gin.Context, payload string) {
	returnDatum(c, "blob", true, "text/plain", []byte(payload), nil)
}

func returnError(c *gin.Context, payload error) {
	returnDatum(c, "blob", false, "text/plain", []byte(payload.Error()), nil)
}