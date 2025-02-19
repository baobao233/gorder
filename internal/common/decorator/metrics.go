package decorator

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MetricClient interface {
	Inc(key string, value int)
}

type queryMetricsDecorator[C, R any] struct {
	client MetricClient       // 把MetricClient传进来，需要用到 Metric，用于记录耗时
	base   QueryHandler[C, R] // 把QueryHandler传进来
}

func (q queryMetricsDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	start := time.Now()
	actionName := strings.ToLower(generateActionName(cmd))
	defer func() {
		end := time.Since(start)
		// 把耗时指标传进 Metric 中，并且采用唯一一次查询的标识
		q.client.Inc(fmt.Sprintf("queries.%s.duration", actionName), int(end.Seconds()))
		if err == nil {
			// 记录成功请求次数
			q.client.Inc(fmt.Sprintf("queries.%s.success", actionName), 1)
		} else {
			// 记录失败请求次数
			q.client.Inc(fmt.Sprintf("queries.%s.failure", actionName), 1)
		}
	}()
	return q.base.Handle(ctx, cmd)
}

type commandMetricsDecorator[C, R any] struct {
	client MetricClient         // 把MetricClient传进来，需要用到 Metric，用于记录耗时
	base   CommandHandler[C, R] // 把QueryHandler传进来
}

func (q commandMetricsDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	start := time.Now()
	actionName := strings.ToLower(generateActionName(cmd))
	defer func() {
		end := time.Since(start)
		// 把耗时指标传进 Metric 中，并且采用唯一一次查询的标识
		q.client.Inc(fmt.Sprintf("command.%s.duration", actionName), int(end.Seconds()))
		if err == nil {
			q.client.Inc(fmt.Sprintf("command.%s.success", actionName), 1)
		} else {
			q.client.Inc(fmt.Sprintf("command.%s.failure", actionName), 1)
		}
	}()
	return q.base.Handle(ctx, cmd)
}
