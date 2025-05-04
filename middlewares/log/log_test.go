package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/niuniumart/sdk/middlewares/nlog"
	"github.com/smartystreets/goconvey/convey"
)

func TestLog(t *testing.T) {
	Init(WithLogLevel("debug"),
		WithFileName("sdk.log"),
		WithMaxSize(100),
		WithMaxBackups(3),
		WithLogPath("./log"),
		WithConsole(true), // 默认不会输出到控制台
	)

	convey.Convey("TestGetLog", t, func() {
		fmt.Println(GetDefaultLogger() == nil, GetDefaultLogger())
	})
	convey.Convey("TestLog", t, func() {
		Infof("sdk test %v", "success")
		Errorf("sdk test %v", "success")
		Warnf("sdk test %v", "success")

		Error("sdk test")
		Info("sdk test")
		Warn("sdk test")

		ctx := context.Background()
		InfoContextf(context.WithValue(ctx, nlog.TraceID, "123132321"), "a is %d", 1)
		ErrorContextf(context.WithValue(ctx, nlog.TraceID, "123132321"), "a is %d", 1)
		WarnContextf(context.WithValue(ctx, nlog.TraceID, "123132321"), "a is %d", 1)
		DebugContextf(context.WithValue(ctx, nlog.TraceID, "123132321"), "a is %d", 1)

		//  Fatalf("sdk test")

	})

}
