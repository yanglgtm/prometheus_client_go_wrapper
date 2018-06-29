package prometheus_client_go_wrapper

import "testing"

var ins *PrometheusWrapper

func init() {
	ins = NewPrometheusWrapper(&Config{App: "test", LogApi: []string{"/test"}})
}

func BenchmarkLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ins.Log("/test", "GET", "200", 0, 0, 0)
	}
}

func BenchmarkRequestLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ins.RequestLog("backend", "/test1", "GET", "200")
	}
}
