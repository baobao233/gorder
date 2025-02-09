package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/logging"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Logger
	base   QueryHandler[C, R] // 把QueryHandler传进来，用于记录
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	body, _ := json.Marshal(cmd) // 因为 cmd 中会有数组，所以在 generateActionName 中大概是打印不出这个结构体的，因此需要转换成 json
	fields := logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	}
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query executed successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute query,err=%v", err)
		}
	}()
	res, err = q.base.Handle(ctx, cmd)
	return res, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Logger
	base   CommandHandler[C, R] // 把QueryHandler传进来，用于记录
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	body, _ := json.Marshal(cmd) // 因为 cmd 中会有数组，所以在 generateActionName 中大概是打印不出这个结构体的，因此需要转换成 json
	fields := logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	}
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Command executed successfully")
		} else {
			logging.Errorf(ctx, fields, "Failed to execute command,err=%v", err)
		}
	}()
	res, err = q.base.Handle(ctx, cmd)
	return res, err
}

// 假设查询是 query.XXXHandler，则返回 XXXHandler
func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
