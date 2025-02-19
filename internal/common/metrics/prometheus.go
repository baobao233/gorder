package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type PrometheusMetricsClient struct {
	registry *prometheus.Registry
}

var dynamicCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dynamic_counter",
	Help: "Count custom keys"},
	[]string{"key"},
)

type PrometheusMetricsConfig struct {
	Host        string
	ServiceName string
}

func NewPrometheusMetricsClient(conf *PrometheusMetricsConfig) *PrometheusMetricsClient {
	client := &PrometheusMetricsClient{}
	client.initPrometheus(conf)
	return client
}

func (p *PrometheusMetricsClient) initPrometheus(conf *PrometheusMetricsConfig) {
	p.registry = prometheus.NewRegistry()
	p.registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// custom collector
	_ = p.registry.Register(dynamicCounter)

	// metadata wrap
	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": conf.ServiceName}, p.registry)

	// export data
	http.Handle("/metrics", promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}))

	// run http
	go func() {
		logrus.Fatalf("failed to start prometheus endpoint,err=%v", http.ListenAndServe(conf.Host, nil))
	}()
}

func (p *PrometheusMetricsClient) Inc(key string, value int) {
	dynamicCounter.WithLabelValues(key).Add(float64(value))
}
