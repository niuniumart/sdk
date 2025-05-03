package ngin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niuniumart/sdk/middlewares/nlog"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	REQUEST_SUCCESS = 0
)

// RequestGetter 获取包体通用结构
type RequestGetter struct {
	RequestID string `json:"RequestId" form:"RequestId"`
	Token     string `json:"token" form:"token"`
	Module    string `json:"module" form:"module"`
}

// DefaultRespGetter 默认返回结构获取器
type DefaultRespGetter struct {
	Code int    `json:"retCode"`
	Msg  string `json:"retMsg"`
}

// GetCode 获取Code属性
func (p *DefaultRespGetter) GetCode() int {
	return p.Code
}

var defaultRespGetterFactory = func() RespGetter {
	return new(DefaultRespGetter)
}

// SetRespGetterFactory 设置返回器工厂
func SetRespGetterFactory(factory RespGetterFactory) {
	respGetterFactory = factory
}

// GetRespGetterFactory 获取返回器工厂
func GetRespGetterFactory() RespGetterFactory {
	return respGetterFactory
}

// RespGetterFactory
type RespGetterFactory func() RespGetter

var respGetterFactory RespGetterFactory

// RespGetter 返回处理器
type RespGetter interface {
	GetCode() int
}

// BodyLogWriter 日志打印器
type BodyLogWriter struct {
	gin.ResponseWriter
	BodyBuf *bytes.Buffer
}

// Write implement body log writer
func (w BodyLogWriter) Write(b []byte) (int, error) {
	//memory copy here!
	w.BodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

/*
*
UrlMetrics
URL_CONFIG_UPDATE_NOTIFY
UrlHeartBeat
*/
const (
	UrlMetrics   = "/metrics"
	UrlHeartBeat = "/heartbeat"
)

// IgnorePaths
var IgnorePaths []string

func init() {
	respGetterFactory = defaultRespGetterFactory
	IgnorePaths = []string{
		UrlMetrics,
		UrlHeartBeat,
	}
}

/*
*
TotalCounterVec
ReqDurationVec
ReqLogicErrorVec
ReqSystemErrorVec
*/
var (
	TotalCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_count",
			Help: "Total number of HTTP requests made",
		},
		[]string{"module", "operation"},
	)
	ReqDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "request_latency",
		Help: "record request latency",
	}, []string{"module", "operation"})
	ReqLogicErrorVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_error_count",
		Help: "Total request error count of the host",
	}, []string{"module", "operation", "code"})
	ReqSystemErrorVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "server_error",
		Help: "Total error count of the request",
	}, []string{"module", "operation"})
)

func init() {
	prometheus.MustRegister(
		TotalCounterVec,
		ReqDurationVec,
		ReqLogicErrorVec,
		ReqSystemErrorVec,
	)
}

// InList func
func InList(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

// GetFmtStr func
func GetFmtStr(data interface{}) string {
	resp, _ := json.Marshal(data)
	respStr := string(resp)
	if respStr == "" {
		respStr = fmt.Sprintf("%+v", data)
	}
	return respStr
}

// DoRequest do request
func DoRequest(uri, method string, param map[string]interface{}) string {
	var reqBody []byte
	if method == http.MethodGet {
		uri = uri + ParseToStr(param)
		u, _ := url.Parse(uri)
		q := u.Query()
		u.RawQuery = q.Encode() //urlencode

		// 构造post请求，json数据以请求body的形式传递
	}
	if method == http.MethodPost {
		reqBody, _ = json.Marshal(param)
	}
	nlog.Infof(context.Background(), "method %s, uri %s, reqBody %s", method, uri, string(reqBody))
	req, err := http.NewRequest(method, uri, bytes.NewReader(reqBody))
	if err != nil {
		nlog.Errorf(context.Background(), "NewRequest err %s", err.Error())
		return err.Error()
	}

	// 初始化响应
	client := &http.Client{Timeout: 3000 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		nlog.Errorf(context.Background(), "NewRequest err %s", err.Error())
		return err.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

// ParseToStr parse map to str
func ParseToStr(mp map[string]interface{}) string {
	values := ""
	if len(mp) == 0 {
		return values
	}
	for key, val := range mp {
		values += "&" + key + "=" + interface2String(val)
	}
	temp := values[1:]
	values = "?" + temp
	return values
}

// interface to string
func interface2String(inter interface{}) string {
	result := ""
	switch inter.(type) {

	case string:
		result = inter.(string)
		break
	case int:
		result = strconv.Itoa(inter.(int))
		break
	case float64:
		strconv.FormatFloat(inter.(float64), 'f', -1, 64)
		break
	}
	return result
}
