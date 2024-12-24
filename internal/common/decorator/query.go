package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// QueryHandler defines a generic type that receives a Query Q,
// and returns a result R
// 对于订单来说我们需要实现这个 handle 方法， 对于不同的 query中不同的参数封装成一个结构体，传入到 Q 中，这样我们还能只能返回啥然后传给 R
// 这样我们只需要 focus 查询什么，返回什么，而不是需要关注具体的实现
type QueryHandler[Q, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// ApplyQueryDecorators 把 logger 传入到 QueryHandler 中,这样子我们就可以得到一个可以记录 log 的 QueryHandler，
// 然后再在 base 中可以再传一个 queryMetricDecorator，因为queryMetricDecorator实现了 handle 所以可以直接传入，
// 从而在一个函数中传入了两个装饰器
func ApplyQueryDecorators[H, R any](handler QueryHandler[H, R], logger *logrus.Entry, metricClient MetricClient) QueryHandler[H, R] {
	return queryLoggingDecorator[H, R]{
		logger: logger,
		base: queryMetricsDecorator[H, R]{
			client: metricClient,
			base:   handler,
		},
	}
}
