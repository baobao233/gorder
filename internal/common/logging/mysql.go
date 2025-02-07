package logging

import (
	"context"
	"github.com/baobao233/gorder/common/util"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	Method   = "method"
	Args     = "args"
	Cost     = "cost"
	Response = "resp"
	Error    = "error"
)

// ArgFormatter 为了让不同的表都能打日志，我们需要通过依赖倒置让想打日志的表都去实现这个接口
type ArgFormatter interface {
	FormatArg() (string, error)
}

func WhenMySQL(ctx context.Context, method string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		Method: method,
		Args:   formatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields[Cost] = time.Since(start).Milliseconds()
		fields[Response] = resp

		if err != nil && (*err) != nil {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields[Error] = err
		}

		logf(ctx, level, fields, "%s", msg)
	}
}

func formatArgs(args []any) string {
	var item []string
	for _, arg := range args {
		item = append(item, formatArg(arg))
	}
	return strings.Join(item, "||")
}

func formatArg(arg any) string {
	var (
		str string
		err error
	)
	defer func() {
		if err != nil {
			str = "unsupported type in formatMySQLArg||err=" + err.Error()
		}
	}()
	switch v := arg.(type) {
	default:
		str, err = util.MarshallString(v)
	case ArgFormatter:
		str, err = v.FormatArg()
	}
	return str
}
