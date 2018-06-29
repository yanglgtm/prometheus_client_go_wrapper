package prometheus_client_go_wrapper

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	App            string
	Idc            string
	LogApi         []string
	LogMethod      []string
	Buckets        []float64
	Objectives     map[float64]float64
	DefaultCollect bool
	// 服务配置
	Service struct {
		ListenPort int
	}
}

type PrometheusWrapper struct {
	c   Config
	reg *prometheus.Registry

	gaugeState       *prometheus.GaugeVec
	histogramLatency *prometheus.HistogramVec
	summaryLatency   *prometheus.SummaryVec

	counterRequests, counterSendBytes  *prometheus.CounterVec
	counterRcvdBytes, counterException *prometheus.CounterVec
}

func (p *PrometheusWrapper) initMonitors() {
	// 请求数
	p.counterRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_requests",
			Help: "number of module requests",
		},
		[]string{"app", "idc", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterRequests)

	// 出口流量
	p.counterSendBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_send_bytes",
			Help: "number of module send bytes",
		},
		[]string{"app", "idc", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterSendBytes)

	// 入口流量
	p.counterRcvdBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_rcvd_bytes",
			Help: "number of module receive bytes",
		},
		[]string{"app", "idc", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterRcvdBytes)

	// 延迟 histogram
	p.histogramLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "histogram_latency",
			Help:    "histogram of module latency",
			Buckets: p.c.Buckets,
		},
		[]string{"app", "idc", "module", "api", "method"},
	)
	p.reg.MustRegister(p.histogramLatency)

	// 延迟 summary
	p.summaryLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "summary_latency",
			Help:       "summary of module latency",
			Objectives: p.c.Objectives,
		},
		[]string{"app", "idc", "module", "api", "method"},
	)
	p.reg.MustRegister(p.summaryLatency)

	// 状态
	p.gaugeState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gauge_state",
			Help: "gauge of app state",
		},
		[]string{"app", "idc", "state"},
	)
	p.reg.MustRegister(p.gaugeState)

	// 异常
	p.counterException = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_exception",
			Help: "number of module exception",
		},
		[]string{"app", "idc", "module", "exception"},
	)
	p.reg.MustRegister(p.counterException)

	if p.c.DefaultCollect {
		p.reg.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
		p.reg.MustRegister(prometheus.NewGoCollector())
	}
}

func (p *PrometheusWrapper) run() {
	go func() {
		http.Handle("/metrics", promhttp.InstrumentMetricHandler(
			p.reg,
			promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})),
		)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.c.Service.ListenPort), nil))
	}()
}

func (p *PrometheusWrapper) inArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (p *PrometheusWrapper) Log(api, method, code string, sendBytes, rcvdBytes, latency float64) {
	if !p.inArray(method, p.c.LogMethod) {
		return
	}
	if !p.inArray(api, p.c.LogApi) {
		return
	}
	p.counterRequests.WithLabelValues(p.c.App, p.c.Idc, "self", api, method, code).Inc()
	if sendBytes > 0 {
		p.counterSendBytes.WithLabelValues(p.c.App, p.c.Idc, "self", api, method, code).Add(sendBytes)
	}
	if rcvdBytes > 0 {
		p.counterRcvdBytes.WithLabelValues(p.c.App, p.c.Idc, "self", api, method, code).Add(rcvdBytes)
	}
	if len(p.c.Buckets) > 0 {
		p.histogramLatency.WithLabelValues(p.c.App, p.c.Idc, "self", api, method).Observe(latency)
	}
	if len(p.c.Objectives) > 0 {
		p.summaryLatency.WithLabelValues(p.c.App, p.c.Idc, "self", api, method).Observe(latency)
	}
}

func (p *PrometheusWrapper) RequestLog(module, api, method, code string) {
	p.counterRequests.WithLabelValues(p.c.App, p.c.Idc, module, api, method, code).Inc()
}

func (p *PrometheusWrapper) SendBytesLog(module, api, method, code string, byte float64) {
	p.counterSendBytes.WithLabelValues(p.c.App, p.c.Idc, module, api, method, code).Add(byte)
}

func (p *PrometheusWrapper) RcvdBytesLog(module, api, method, code string, byte float64) {
	p.counterRcvdBytes.WithLabelValues(p.c.App, p.c.Idc, module, api, method, code).Add(byte)
}

func (p *PrometheusWrapper) HistogramLatencyLog(module, api, method string, latency float64) {
	p.histogramLatency.WithLabelValues(p.c.App, p.c.Idc, module, api, method).Observe(latency)
}

func (p *PrometheusWrapper) SummaryLatencyLog(module, api, method string, latency float64) {
	p.summaryLatency.WithLabelValues(p.c.App, p.c.Idc, module, api, method).Observe(latency)
}

func (p *PrometheusWrapper) ExceptionLog(module, exception string) {
	p.counterException.WithLabelValues(p.c.App, p.c.Idc, module, exception).Inc()
}

func (p *PrometheusWrapper) StateLog(state string, value float64) {
	p.gaugeState.WithLabelValues(p.c.App, p.c.Idc, state).Set(value)
}

func NewPrometheusWrapper(conf *Config) *PrometheusWrapper {
	if conf.App == "" {
		panic("missing App config")
	}
	if conf.Idc == "" {
		conf.Idc = "none"
	}
	if len(conf.LogMethod) == 0 {
		conf.LogMethod = []string{"GET", "POST"}
	}
	if conf.Service.ListenPort == 0 {
		conf.Service.ListenPort = 8080
	}

	for k, v := range conf.LogApi {
		conf.LogApi[k] = strings.ToLower(v)
	}

	w := &PrometheusWrapper{
		c:   *conf,
		reg: prometheus.NewRegistry(),
	}

	w.initMonitors()
	w.run()

	return w
}
