package main

//
//import (
//	"bytes"
//	"encoding/json"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/collectors"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
//	"log"
//	"math/rand"
//	"net/http"
//	"time"
//)
//
//const (
//	testAddr = "localhost:9123"
//)
//
//var httpStatusCodeCounter = prometheus.NewCounterVec(
//	prometheus.CounterOpts{
//		Name: "http_status_code_counter",
//		Help: "Count HTTP status codes",
//	},
//	[]string{"status_code"},
//)
//
//func main() {
//	go produceData()
//	reg := prometheus.NewRegistry()
//	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "demo-service"}, reg).MustRegister(
//		collectors.NewGoCollector(),
//		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
//		httpStatusCodeCounter,
//	)
//
//	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
//	http.HandleFunc("/", sendMetricsHandler)
//	log.Fatal(http.ListenAndServe(testAddr, nil))
//}
//
//func sendMetricsHandler(writer http.ResponseWriter, r *http.Request) {
//	var req request
//	defer func() {
//		httpStatusCodeCounter.WithLabelValues(req.StatusCode).Inc()
//		log.Printf("add 1 to %s", req.StatusCode)
//	}()
//	_ = json.NewDecoder(r.Body).Decode(&req)
//	log.Printf("recevie req %+v", req)
//	_, _ = writer.Write([]byte(req.StatusCode))
//}
//
//type request struct {
//	StatusCode string
//}
//
//func produceData() {
//	codes := []string{"404", "304", "402", "500", "200"}
//	for {
//		code := codes[rand.Intn(len(codes))]
//		body, _ := json.Marshal(&request{
//			StatusCode: code,
//		})
//		requestBody := bytes.NewBuffer(body)
//		http.Post("http://"+testAddr, "application/json", requestBody)
//		log.Printf("send code %s to %s", requestBody.String(), testAddr)
//		time.Sleep(2 * time.Second)
//	}
//}
