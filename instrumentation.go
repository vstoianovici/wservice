package wservice

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           WalletService
}

// NewInstrumenting exported to be accessable from outside the package (from main)
func NewInstrumenting(requestCount metrics.Counter, requestLatency metrics.Histogram, next WalletService) WalletService {
	return &instrumentingMiddleware{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		next:           next,
	}
}

func (mw instrumentingMiddleware) GetTable(s string) (output []string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "getTable", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.next.GetTable(s)
	return
}

func (mw instrumentingMiddleware) DoTransfer(s string, t string, v string) (output string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "doTransfers", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.next.DoTransfer(s, t, v)
	return
}
