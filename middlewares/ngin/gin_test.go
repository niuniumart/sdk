package ngin

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niuniumart/sdk/middlewares/nlog"
	"github.com/smartystreets/goconvey/convey"
)

func TestCreateGin(t *testing.T) {
	convey.Convey("TestCreateGin", t, func() {
		nlog.Infof(context.Background(), "just test log", 5, 666)
		engine := CreateGin()
		engine.POST("/reverse", Reverse)
		engine.GET("/panic", MustPanic)
		engine.GET("/logicerr", LogicError)
		go func() {
			engine.Run(":30001")
			time.Sleep(20 * time.Second)
		}()
		var url = "http://127.0.0.1:30001/reverse"
		param := make(map[string]interface{})
		param["abc"] = "xxx"
		//		requestid.Set("heiheiheihei")
		resp := DoRequest(url, http.MethodPost, param)
		fmt.Printf("resp %s\n", resp)

		url = "http://127.0.0.1:30001/logicerr"
		traceID := "1234567890"
		param = make(map[string]interface{})
		param["abc"] = "xxx"
		param["traceID"] = traceID
		//		requestid.Set("heiheiheihei")
		resp = DoRequest(url, http.MethodGet, param)
		fmt.Printf("resp %s\n", resp)

		url = "http://127.0.0.1:30001/panic"
		param = make(map[string]interface{})
		param["abc"] = "xxx"
		//		requestid.Set("heiheiheihei")
		resp = DoRequest(url, http.MethodGet, param)
		fmt.Printf("resp %s\n", resp)
	})
}

func Pong(c *gin.Context) {
	fmt.Println("pong")
	c.JSON(http.StatusOK, "pong")
}

func Reverse(c *gin.Context) {
	c.JSON(http.StatusOK, "into")
}

func MustPanic(c *gin.Context) {
	panic(nil)
}

func LogicError(c *gin.Context) {
	traceID := c.Request.Header.Get(nlog.TraceID)
	ctx := context.Background()
	ctx = context.WithValue(ctx, nlog.TraceID, traceID)
	nlog.Infof(ctx, "LogicError !!!!!")
	var resp Resp
	resp.RetCode = 10005
	resp.RetMsg = "logic error"
	c.JSON(http.StatusOK, &resp)
}

type Resp struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
}
