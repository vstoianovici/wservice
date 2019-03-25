package wservice

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
)

// The instrumentation middleware amends the wallet service and any other layer that wraps it with an instrumentation layer

// instrumentingMiddleware is the type of the wrapper around the core service and any other functionality layers
type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           WalletService
}

// NewInstrumenting is how the instrumenting middleware (instrumentingMiddleware struct) is constructed (the function is exported so it can be used from outside the package)
func NewInstrumenting(requestCount metrics.Counter, requestLatency metrics.Histogram, next WalletService) WalletService {
	return &instrumentingMiddleware{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		next:           next,
	}
}

// GetTable function is implemented for the instrumenting layer as the request traverses through the instrumenting layer down to the next layer
func (mw instrumentingMiddleware) GetTable(s string) (output []string, err error) {
	// Incremement instrumenting counters and determine latency
	defer func(begin time.Time) {
		lvs := []string{"method", "getTable", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	// The function calls the next layer down
	output, err = mw.next.GetTable(s)
	return
}

// DoTransfer function is implemented for the instrumenting layer as the request traverses through the instrumenting layer down to the next layer
func (mw instrumentingMiddleware) DoTransfer(s string, t string, v string) (output string, err error) {
	// Incremement instrumenting counters and determine latency
	defer func(begin time.Time) {
		lvs := []string{"method", "doTransfers", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	// The function calls the next layer down
	output, err = mw.next.DoTransfer(s, t, v)
	return
}
