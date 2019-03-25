package wservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Transports exposes the wallet service API to the network via JSON over HTTP. As a future extension gRPC could also be introduced for inter-service commuincation.

// NewHTTPTransport creates a new JSON over HTTP transport
func NewHTTPTransport(svc WalletService) http.Handler {
	// define a way to service a request for the TransfersEndpoint
	transfersHandler := httptransport.NewServer(
		MakeTransfersEndpoint(svc),
		DecodeTransfersRequest,
		EncodeResponse,
	)
	// define a way to service a request for the AccountsEndpoint
	accountsHandler := httptransport.NewServer(
		MakeAccountsEndpoint(svc),
		DecodeAccountsRequest,
		EncodeResponse,
	)
	// define a way to service a request for the submitTransferEndpoint
	submitTransferHandler := httptransport.NewServer(
		MakeSubmitTransferEndpoint(svc),
		DecodeSubmitTransferRequest,
		EncodeResponse,
	)
	// Define a new router that will handle API endpoints for each of the previously defined handlers and for metrics
	r := mux.NewRouter()
	r.Handle("/transfers", transfersHandler)
	r.Handle("/accounts", accountsHandler)
	r.Handle("/submittransfer", submitTransferHandler)
	r.Handle("/metrics", promhttp.Handler())
	// Return the router
	return r
}

// DecodeTransfersRequest exported to be accessable from outside the package (from main)
func DecodeTransfersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Method == http.MethodGet {
		return nil, nil
	}
	var ErrVerb = errors.New("err: Verb can only be \"GET\" for endpoint \"/transfers\"")
	return nil, ErrVerb
}

// DecodeAccountsRequest exported to be accessable from outside the package (from main)
func DecodeAccountsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Method == http.MethodGet {
		return nil, nil
	}
	var ErrVerb = errors.New("err: Verb can only be \"GET\" for endpoint \"/accounts\"")
	return nil, ErrVerb
}

// DecodeSubmitTransferRequest exported to be accessable from outside the package (from main)
func DecodeSubmitTransferRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Method != http.MethodPost {
		var ErrVerb = errors.New("err: Verb can only be \"POST\" for endpoint \"/submittransfer\"")
		return nil, ErrVerb
	}

	var request submitTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

// EncodeResponse exported to be accessable from outside the package (from main)
func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
