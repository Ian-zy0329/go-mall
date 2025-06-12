package middleware

import (
	"bytes"
	"github.com/Ian-zy0329/go-mall/common/logger"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

func StartTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.Request.Header.Get("traceid")
		pSpanId := c.Request.Header.Get("spanid")
		spanId := util.GenerateSpanID(c.Request.RemoteAddr)
		if traceId == "" {
			traceId = spanId
		}
		c.Set("traceid", traceId)
		c.Set("spanid", spanId)
		c.Set("pspanid", pSpanId)
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqBody, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		start := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		accessLog(c, "access_start", time.Since(start), reqBody, nil)
		defer func() {
			accessLog(c, "access_end", time.Since(start), reqBody, blw.body.String())
		}()
		c.Next()
		return
	}
}

func accessLog(c *gin.Context, accessType string, dur time.Duration, body []byte, dataOut interface{}) {
	req := c.Request
	bodyStr := string(body)
	query := req.URL.RawQuery
	path := req.URL.Path
	logger.New(c).Info("AccessLog",
		"type", accessType,
		"ip", c.ClientIP(),
		"method", req.Method,
		"path", path,
		"query", query,
		"body", bodyStr,
		"output", dataOut,
		"time(ms)", int64(dur/time.Millisecond),
	)
}

func GinPanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.New(c).Error("http request broken pipe", "path", c.Request.URL.Path, "error", err, "request", string(httpRequest))
					c.Error(err.(error))
					c.Abort()
					return
				}
				logger.New(c).Error("http_request_panic", "path", c.Request.URL.Path, "error", err, "request", string(httpRequest), "stack", string(debug.Stack()))
				c.AbortWithError(http.StatusInternalServerError, err.(error))
			}
		}()
		c.Next()
	}
}
