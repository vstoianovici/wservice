package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	wservice "github.com/vstoianovici/wservice"
)

// Transports expose the service to the network. In this first example we utilize JSON over HTTP.
func main() {
	logger := createLogger()
	startLogger := log.With(logger, "tag", "start")
	startLogger.Log("msg", "created logger")

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
	svc, port, err = wservice.NewService()
	if err != nil {
		startLogger.Log("msg", "failed to connect to database", "err", err)
		os.Exit(1)
	}
	sPortNumber := ":" + strconv.Itoa(port)
	svc = wservice.NewLogging(logger, svc)
	svc = wservice.NewInstrumenting(requestCount, requestLatency, svc)

	transfersHandler := httptransport.NewServer(
		wservice.MakeTransfersEndpoint(svc),
		wservice.DecodeTransfersRequest,
		wservice.EncodeResponse,
	)

	accountsHandler := httptransport.NewServer(
		wservice.MakeAccountsEndpoint(svc),
		wservice.DecodeAccountsRequest,
		wservice.EncodeResponse,
	)

	submitTransferHandler := httptransport.NewServer(
		wservice.MakeSubmitTransferEndpoint(svc),
		wservice.DecodeSubmitTransferRequest,
		wservice.EncodeResponse,
	)

	http.Handle("/transfers", transfersHandler)
	http.Handle("/accounts", accountsHandler)
	http.Handle("/submittransfer", submitTransferHandler)
	http.Handle("/metrics", promhttp.Handler())
	startLogger.Log("msg", "GET Acounts Table here: http://127.0.0.1:8080/accounts")
	startLogger.Log("msg", "GET Transfers Table here: http://127.0.0.1:8080/transfers")
	startLogger.Log("msg", "GET Metrics & Instrumentation here: http://127.0.0.1:8080/metrics")
	startLogger.Log("msg", "HTTP serving", "addr", sPortNumber)

	startLogger.Log(http.ListenAndServe(sPortNumber, nil))
}

func createLogger() log.Logger {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	return log.With(logger, "time", log.DefaultTimestampUTC())
}
