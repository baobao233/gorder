package middlewares

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

func RequestLog(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIn(c, l)
		defer requestOut(c, l)
		c.Next()
	}
}

func requestIn(c *gin.Context, l *logrus.Entry) {
	c.Set("request_start", time.Now())
	body := c.Request.Body
	bodyBytes, _ := io.ReadAll(body)
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // 把 io.ReadAll 读出来的内容写回去，如果不写回去会把 body取出来变成 EOF 错误

	var compactJson bytes.Buffer
	_ = json.Compact(&compactJson, bodyBytes) // 去除 JSON 那些不影响数据表示的空格字符
	l.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"time": time.Now().Unix(),
		"args": compactJson.String(),
		"from": c.RemoteIP(),
		"uri":  c.Request.RequestURI,
	}).Info("__request_in")
}

func requestOut(c *gin.Context, l *logrus.Entry) {
	response, _ := c.Get("response") // response.go 中设置的
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)
	l.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"proc_time_ms": time.Since(startTime).Milliseconds(),
		"response":     response,
	}).Info("__request_out")
}
