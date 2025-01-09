package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R] // 把QueryHandler传进来，用于记录
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	body, _ := json.Marshal(cmd) // 因为 cmd 中会有数组，所以在 generateActionName 中大概是打印不出这个结构体的，因此需要转换成 json
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	})
	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query executed successfully ")
		} else {
			logger.Error("Failed to execute query", err)
		}
	}()
	res, err = q.base.Handle(ctx, cmd)
	return res, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   CommandHandler[C, R] // 把QueryHandler传进来，用于记录
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	body, _ := json.Marshal(cmd) // 因为 cmd 中会有数组，所以在 generateActionName 中大概是打印不出这个结构体的，因此需要转换成 json
	logger := q.logger.WithFields(logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	})
	logger.Debug("Executing command")
	defer func() {
		if err == nil {
			logger.Info("Command executed successfully ")
		} else {
			logger.Error("Failed to execute command", err)
		}
	}()
	res, err = q.base.Handle(ctx, cmd)
	return res, err
}

// 假设查询是 query.XXXHandler，则返回 XXXHandler
func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
