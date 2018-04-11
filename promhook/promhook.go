package promhook

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twitchtv/twirp"
)

var reqStartTimestampKey = new(int)

func markReqStart(ctx context.Context) context.Context {
	return context.WithValue(ctx, reqStartTimestampKey, time.Now())
}

func getReqStart(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(reqStartTimestampKey).(time.Time)
	return t, ok
}

// NewServerHooks initializes twirp server hooks that record prometheus metrics
// of twirp operations
func NewServerHooks(ns string) *twirp.ServerHooks {

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "trpc_requests_total",
			Help:      "Counter of total requests received.",
		},
		[]string{"package", "service", "method"},
	)

	responseCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "trpc_responses_total",
			Help:      "Counter of total responses sent.",
		},
		[]string{"package", "service", "method", "status"},
	)
	// Is there a way to count the number of results sent?
	// resultCount := prometheus.NewSummaryVec(
	// 	prometheus.SummaryOpts{
	// 		Namespace: ns,
	// 		Name:      "trpc_result_total",
	// 		Help:      "The sum of results for a method.",
	// 	},
	// 	[]string{}, // no fields here
	// )

	requestLatency := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "trpc_request_latency",
			Help:      "Total duration of requests in microseconds.",
		},
		[]string{"package", "service", "method", "status"},
	)

	var registerer = prometheus.DefaultRegisterer
	registerer.MustRegister(requestCount)
	registerer.MustRegister(responseCount)
	registerer.MustRegister(requestLatency)

	hooks := &twirp.ServerHooks{}

	hooks.RequestReceived = func(ctx context.Context) (context.Context, error) {
		ctx = markReqStart(ctx)
		return ctx, nil
	}

	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
		pkg, ok := twirp.PackageName(ctx)
		service, ok := twirp.ServiceName(ctx)
		method, ok := twirp.MethodName(ctx)
		if !ok {
			return ctx, nil
		}

		requestCount.WithLabelValues(
			sanitize(pkg),
			sanitize(service),
			sanitize(method),
		).Add(1)
		return ctx, nil
	}

	hooks.ResponseSent = func(ctx context.Context) {
		pkg, _ := twirp.PackageName(ctx)
		service, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)
		status, _ := twirp.StatusCode(ctx)

		responseCount.WithLabelValues(
			sanitize(pkg),
			sanitize(service),
			sanitize(method),
			status,
		).Add(1)

		if start, ok := getReqStart(ctx); ok {
			dur := time.Now().Sub(start).Seconds()
			requestLatency.WithLabelValues(
				sanitize(pkg),
				sanitize(service),
				sanitize(method),
				status,
			).Observe(dur)
		}
	}
	return hooks
}

func sanitize(s string) string {
	return strings.Map(sanitizeRune, s)
}

func sanitizeRune(r rune) rune {
	switch {
	case 'a' <= r && r <= 'z':
		return r
	case '0' <= r && r <= '9':
		return r
	case 'A' <= r && r <= 'Z':
		return r
	default:
		return '_'
	}
}
