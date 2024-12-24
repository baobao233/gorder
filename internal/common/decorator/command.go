package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// CommandHandler defines a generic type that receives a command C,
// and returns a result R
// 对于订单来说我们需要实现这个 handle 方法， 对于不同的 command 中不同的参数封装成一个结构体，传入到 C 中，这样我们还能只能返回啥然后传给 R
// 这样我们只需要 focus 查询什么，返回什么，而不是需要关注具体的实现
type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// ApplyCommandDecorators 把 logger 传入到 CommandHandler 中,这样子我们就可以得到一个可以记录 log 的 CommandHandler，
func ApplyCommandDecorators[C, R any](handler CommandHandler[C, R], logger *logrus.Entry, metricClient MetricClient) CommandHandler[C, R] {
	return queryLoggingDecorator[C, R]{
		logger: logger,
		base: queryMetricsDecorator[C, R]{
			client: metricClient,
			base:   handler,
		},
	}
}
