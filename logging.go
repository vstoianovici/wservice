package wservice

import (
	"time"

	"github.com/go-kit/kit/log"
)

// The logging middleware amends the wallet service with a logger

// loggingMiddleware is the type of the wrapper around the core service and any other functionality layers
type loggingMiddleware struct {
	logger log.Logger
	next   WalletService
}

// NewLogging is how the logging middleware (loggingMiddleware struct) is constructed (the function is exported so it can be used from outside the package)
func NewLogging(logger log.Logger, next WalletService) WalletService {
	return &loggingMiddleware{
		logger: logger,
		next:   next,
	}
}

// GetTable function is implemented for logging layer as the request traverses through the logging layer down to the next layer
func (mw loggingMiddleware) GetTable(s string) (output []string, err error) {
	// Log everything that the function sees in the provided format
	defer func(begin time.Time) {
		status := func(in []string) string {
			if len(in) != 0 {
				return "success"
			}
			return "no table"
		}(output)
		_ = mw.logger.Log(
			"method", "getTables",
			"input", s,
			//"output", strings.Join(output, ","),
			"output", status,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	// The function calls the next layer down
	output, err = mw.next.GetTable(s)
	return
}

// DoTransfer function is implemented for the logging layer as the request traverses through the logging layer down to the next layer
func (mw loggingMiddleware) DoTransfer(s string, t string, v string) (output string, err error) {
	// Log everything that the function sees in the provided format
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "doTransfer",
			"input", "From "+s+" to "+t+" amount "+v,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	// The function calls the next layer down
	output, err = mw.next.DoTransfer(s, t, v)
	return
}
