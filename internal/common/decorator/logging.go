package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R] // 把QueryHandler传进来，用于记录
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	})
	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query executed successfully ")
		} else {
			logger.Error("Failed to execute query ", err)
		}
	}()
	return q.base.Handle(ctx, cmd)
}

// 假设查询是 query.XXXHandler，则返回 XXXHandler
func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
