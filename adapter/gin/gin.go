package gin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	pw "github.com/itsmikej/prometheus_client_go_wrapper"
)

type AdapterGin struct {
	prom *pw.PrometheusWrapper
}

func NewAdapterGin(p *pw.PrometheusWrapper) *AdapterGin {
	return &AdapterGin{prom: p}
}

func (a *AdapterGin) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b := time.Now()
		ctx.Next()
		latency := float64(time.Now().Sub(b).Nanoseconds() / 1000000)
		a.prom.Log(ctx.Request.URL.Path, ctx.Request.Method, fmt.Sprintf("%d", ctx.Writer.Status()), float64(ctx.Writer.Size()), 0, latency)
	}
}
