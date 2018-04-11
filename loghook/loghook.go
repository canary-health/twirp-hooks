package loghook

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/twitchtv/twirp"
)

/// WORK IN PROGRESS DO NOT USE!!

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

		return ctx, nil
	}

	hooks.ResponseSent = func(ctx context.Context) {
		pkg, _ := twirp.PackageName(ctx)
		service, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)
		status, _ := twirp.StatusCode(ctx)

		if start, ok := getReqStart(ctx); ok {
			dur := time.Now().Sub(start).Seconds()

		}
		svr := ctx.Value(int32(2))
		w := log.NewSyncWriter(os.Stderr)
		logger := log.NewLogfmtLogger(w)

		defer func(begin time.Time) {
			_ = logger.Log(
				"method", sanitize(method),
				"ctx", ctx,
				"output", svr,
				// "err", err,
				"took", time.Since(begin),
			)
		}(time.Now())
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
