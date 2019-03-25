package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	wservice "github.com/vstoianovici/wservice"
)

func main() {
	// Define a cutom logger with the specified format
	logger := createLogger()
	startLogger := log.With(logger, "tag", "start")
	startLogger.Log("msg", "created logger")

	// define instrumenting entities to be diplayed and accounted for
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "vlad_group",
		Subsystem: "funds_transfer_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "vlad_group",
		Subsystem: "funds_transfer_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	var svc wservice.WalletService
	var err error
	var port int
	// Create a new wallet service and get a post where to listen and serve
	svc, port, err = wservice.NewService()
	// In case of any issues return the error to the log
	if err != nil {
		startLogger.Log("msg", "failed to connect to database", "err", err)
		os.Exit(1)
	}
	sPortNumber := ":" + strconv.Itoa(port)
	// Add a layer of logging on top of the core wallet service
	svc = wservice.NewLogging(logger, svc)
	// Add a layer of instrumenting on top of the core wallet service
	svc = wservice.NewInstrumenting(requestCount, requestLatency, svc)
	// Create a new HTTP Transport layer for the wallet service to serve its API
	httpTransport := wservice.NewHTTPTransport(svc)
	// Add some informational logging messages
	startLogger.Log("msg", "GET Acounts Table here: http://127.0.0.1:8080/accounts")
	startLogger.Log("msg", "GET Transfers Table here: http://127.0.0.1:8080/transfers")
	startLogger.Log("msg", "GET Metrics & Instrumentation here: http://127.0.0.1:8080/metrics")
	startLogger.Log("msg", "HTTP serving", "addr", port)
	// Start the server and log whatever it has to say
	startLogger.Log(http.ListenAndServe(sPortNumber, httpTransport))
}

// createLogger implements the disred log format
func createLogger() log.Logger {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	return log.With(logger, "time", log.DefaultTimestampUTC())
}
