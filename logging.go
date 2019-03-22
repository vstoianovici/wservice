package wservice

import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	logger log.Logger
	next   WalletService
}

// NewLogging is exported to be used in main
func NewLogging(logger log.Logger, next WalletService) WalletService {
	return &loggingMiddleware{
		logger: logger,
		next:   next,
	}
}

func (mw loggingMiddleware) GetTable(s string) (output []string, err error) {
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

	output, err = mw.next.GetTable(s)
	return
}

// DoTransfer exported to be accessable from outside the package (from main)
func (mw loggingMiddleware) DoTransfer(s string, t string, v string) (output string, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "doTransfer",
			"input", "From "+s+" to "+t+" amount "+v,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.next.DoTransfer(s, t, v)
	return
}
