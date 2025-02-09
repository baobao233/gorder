package logging

import (
	"context"
	"github.com/baobao233/gorder/common/tracing"
	"github.com/rifflock/lfshook"
	"os"
	"strconv"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

/*
该文件用于对日志进行标准化的输出，看起来美观
*/

// 要么用 logging.Infof, logging.Warnf...
// 或者直接加 hook，用 logrus.Infof...

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel) // 开发环境中设置为 debug 模式
	setOutput(logrus.StandardLogger())
	logrus.AddHook(&traceHook{})
}

func setOutput(logger *logrus.Logger) {
	var (
		folder    = "./log/"
		filePath  = "app.log"
		errorPath = "error.log"
	)

	if err := os.MkdirAll(folder, 0750); err != nil && !os.IsExist(err) {
		if os.IsExist(err) {
			err = os.RemoveAll(folder)
			if err != nil {
				panic(err)
			}
		}
		panic(err)
	}
	file, err := os.OpenFile(folder+filePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	_, err = os.OpenFile(folder+errorPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(file)

	rotateInfo, err := rotatelogs.New(
		folder+filePath+".%Y%m%d",
		rotatelogs.WithLinkName("app.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(1*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	rotateError, err := rotatelogs.New(
		folder+errorPath+".%Y%m%d",
		rotatelogs.WithLinkName("errors.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(1*time.Hour),
	)
	if err != nil {
		panic(err)
	}

	rotateMap := lfshook.WriterMap{
		logrus.InfoLevel:  rotateInfo,
		logrus.DebugLevel: rotateInfo,
		logrus.WarnLevel:  rotateError,
		logrus.ErrorLevel: rotateError,
		logrus.FatalLevel: rotateError,
		logrus.PanicLevel: rotateError,
	}
	logrus.AddHook(lfshook.NewHook(rotateMap, &logrus.JSONFormatter{
		TimestampFormat: time.DateTime,
	}))
}

func SetFormatter(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})
	// 如果是在本地模式中，强制格式化
	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		logrus.SetFormatter(&prefixed.TextFormatter{
			//ForceColors:     true,
			//ForceFormatting: true,
			//TimestampFormat: time.RFC3339,
		})
	}
}

// 包内使用的 logf
func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args...)
}

func InfofWithTime(ctx context.Context, fields logrus.Fields, start time.Time, format string, args ...any) {
	fields[Cost] = time.Since(start)
	Infof(ctx, fields, format, args...)
}

func Infof(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Infof(format, args...)
}

func Errorf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Errorf(format, args...)
}

func Warnf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Warnf(format, args...)
}

func Panicf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Panicf(format, args...)
}

type traceHook struct{}

// Levels 在哪些 level 需要调用这些hook
func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 调用这个 hook
func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["hook_trace"] = tracing.TraceID(entry.Context)
		entry = entry.WithTime(time.Now())
	}
	return nil
}
