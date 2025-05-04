package ngin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niuniumart/sdk/middlewares/nlog"
)

// Cors cors处理
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

const MAX_PRINT_BODY_LEN = 1024

// writer, use for log
type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

// Write write func
func (w bodyLogWriter) Write(b []byte) (int, error) {
	//memory copy here!
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

var ignoreReqLogUrlDic, ignoreRespLogUrl map[string]int

// init
func init() {
	ignoreReqLogUrlDic = make(map[string]int)
	ignoreRespLogUrl = make(map[string]int)
}

// RegisterIgnoreLogUrl func registerIgnoreLogUrl url
func RegisterIgnoreLogUrl(url string) {
	ignoreReqLogUrlDic[url] = 1
	ignoreRespLogUrl[url] = 1
}

// RegisterIgnoreReqLogUrl register ignore req log url/
func RegisterIgnoreReqLogUrl(url string) {
	ignoreReqLogUrlDic[url] = 1
}

// RegisterIgnoreRespLogUrl register ignore resp log url/
func RegisterIgnoreRespLogUrl(url string) {
	ignoreRespLogUrl[url] = 1
}

// InfoLog func infoLog
func InfoLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if InList(c.Request.URL.Path, IgnorePaths) {
			return
		}
		beginTime := time.Now()
		// ***** 1. get request body ****** //
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close() //  must close
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		// ***** 2. set requestID for goroutine ctx ****** //
		requestID := c.Request.Header.Get(nlog.TraceID)
		if requestID == "" {
			requestID = "empty" + uuid.New().String()
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, nlog.TraceID, requestID)
		c.Request.Header.Set(nlog.TraceID, requestID)
		if _, ok := ignoreReqLogUrlDic[c.Request.URL.Path]; !ok {
			nlog.Infof(ctx, "Req Url: %s %+v,[Body]:%s; [Header]:%s", c.Request.Method, c.Request.URL,
				string(body), GetFmtStr(c.Request.Header))
		}
		// ***** 3. set resp writer ****** //
		blw := BodyLogWriter{BodyBuf: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		// ***** 4. do Next ****** //
		c.Next()
		// ***** 5. log resp body ****** //
		strBody := strings.Trim(blw.BodyBuf.String(), "\n")
		if len(strBody) > MAX_PRINT_BODY_LEN {
			strBody = strBody[:(MAX_PRINT_BODY_LEN - 1)]
		}
		// ***** 6. judge logic error ****** //
		getterFactory := GetRespGetterFactory()
		rspGetter := getterFactory()
		//var rspGetter utils.ResponseGetter
		json.Unmarshal(blw.BodyBuf.Bytes(), &rspGetter)
		if rspGetter.GetCode() != REQUEST_SUCCESS {
			ReqLogicErrorVec.WithLabelValues("", c.Request.URL.Path,
				fmt.Sprintf("%d", rspGetter.GetCode())).Inc()
		}
		if _, ok := ignoreRespLogUrl[c.Request.URL.Path]; !ok {
			nlog.Infof(context.Background(), "Url: %+v, cost %v Resp Body %s", c.Request.URL,
				time.Since(beginTime), strBody)
		}
		duration := float64(time.Since(beginTime)) / float64(time.Second)
		nlog.Infof(ctx, "ReqPath[%s]-Duration[%g]", c.Request.URL.Path, duration)
		ReqDurationVec.WithLabelValues("", c.Request.URL.Path).Observe(duration)
	}
}

// return max(a, b)
func maxCode(a, b int) string {
	if a > b {
		return fmt.Sprintf("%d", a)
	}
	return fmt.Sprintf("%d", b)
}

// body writer
type bodyWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

// Write func write
func (w bodyWriter) Write(b []byte) (int, error) {
	//memory copy here!
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

// PanicRecover func panicRecover
func PanicRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				nlog.Errorf(context.Background(), "In PanicRecover,Error:%s", err)
				var rg RequestGetter
				body, _ := ioutil.ReadAll(c.Request.Body)
				err := json.Unmarshal(body, &rg)
				if err != nil {
					nlog.Warnf(context.Background(), "Req Body json unmarshal requestID err %s", err.Error())
				}
				ReqSystemErrorVec.WithLabelValues(rg.Module, c.Request.URL.Path).Inc()
				//打印调用栈信息
				debug.PrintStack()
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				stackInfo := fmt.Sprintf("%s", buf[:n])
				nlog.Errorf(context.Background(), "panic stack info %s\n", stackInfo)
				/*blw := bodyWriter{bodyBuf: bytes.NewBufferString(""), ResponseWriter: c.Writer}
				c.Writer = blw*/
				c.JSON(http.StatusOK, *BuildFailResp(nil))
				return
			}
		}()
		c.Next()
	}
}
