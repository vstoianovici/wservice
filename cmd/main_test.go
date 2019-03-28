package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	wservice "github.com/vstoianovici/wservice"
)

func TestExternalServiceCreation(t *testing.T) {
	var i int
	svc, port, err := wservice.NewService()
	assert.IsType(t, port, i)
	assert.NoError(t, err)
	status, err := svc.DoTransfer("bob123", "alice456", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	status, err = svc.DoTransfer("alice456", "bob123", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	vSlice, err := svc.GetTable("Accounts")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Transfers")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("someOtherTable")
	assert.NotContains(t, vSlice, "[]")
	logger := createLogger()
	svc = wservice.NewLogging(logger, svc)
	status, err = svc.DoTransfer("bob123", "alice456", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	status, err = svc.DoTransfer("alice456", "bob123", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Accounts")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Transfers")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("someOtherTable")
	assert.NotContains(t, vSlice, "[]")
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
	svc = wservice.NewInstrumenting(requestCount, requestLatency, svc)
	status, err = svc.DoTransfer("bob123", "alice456", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	status, err = svc.DoTransfer("alice456", "bob123", "1")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Accounts")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Transfers")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("someOtherTable")
	assert.NotContains(t, vSlice, "[]")
	portNumber := strconv.Itoa(port)
	var f *mux.Router
	httpTransport := wservice.NewHTTPTransport(svc)
	assert.IsType(t, httpTransport, f)
	http.ListenAndServe(portNumber, httpTransport)
	request, _ := http.NewRequest("GET", "/accounts", nil)
	response := httptest.NewRecorder()
	httpTransport.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
}
