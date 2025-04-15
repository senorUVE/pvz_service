package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"path", "method", "status"},
	)
	ResponseDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_duration_seconds",
			Help: "Response time",
		},
		[]string{"path", "method"},
	)

	PVZCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total created PVZ",
		},
	)
	ReceptionsCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total receptions created",
		},
	)
	ProductsAdded = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total added products",
		},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal, ResponseDur)
	prometheus.MustRegister(PVZCreated, ReceptionsCreated, ProductsAdded)
}

func PrometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTime := time.Now()
		err := next(c)
		duration := time.Since(startTime)

		path := c.Path()
		if path == "" {
			path = c.Request().URL.Path
		}
		status := c.Response().Status
		method := c.Request().Method

		RequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		ResponseDur.WithLabelValues(path, method).Observe(duration.Seconds())

		return err
	}
}

func IncPvzCreated() {
	PVZCreated.Inc()
}

func IncReceptionsCreated() {
	ReceptionsCreated.Inc()
}

func IncProductsAdded() {
	ProductsAdded.Inc()
}

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
