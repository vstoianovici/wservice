package wservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

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

// For each method, we define request and response structs
type transfersRequest struct {
	S string `json:"s"`
}

type transfersResponse struct {
	V   []string `json:"v"`
	Err string   `json:"err,omitempty"` // errors don't define JSON marshaling
}

type accountsRequest struct {
	S string `json:"s"`
}

type accountsResponse struct {
	V   []string `json:"v"`
	Err string   `json:"err,omitempty"` // errors don't define JSON marshaling
}

type submitTransferRequest struct {
	FromAccount string `json:"from"`
	ToAccount   string `json:"to"`
	Amount      string `json:"amount"`
}

type submitTransferResponse struct {
	V   string `json:"result"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

// MakeTransfersEndpoint exported to be accessable from outside the package (from main)
func MakeTransfersEndpoint(svc WalletService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		//req := request.(transfersRequest)
		v, err := svc.GetTable("Transfers")
		if err != nil {
			return transfersResponse{v, err.Error()}, nil
		}
		return transfersResponse{v, ""}, nil
	}
}

// MakeAccountsEndpoint exported to be accessable from outside the package (from main)
func MakeAccountsEndpoint(svc WalletService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		//req := request.(accountsRequest)
		v, err := svc.GetTable("Accounts")
		if err != nil {
			return accountsResponse{v, err.Error()}, nil
		}
		return accountsResponse{v, ""}, nil
	}
}

// MakeSubmitTransferEndpoint exported to be accessable from outside the package (from main)
func MakeSubmitTransferEndpoint(svc WalletService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(submitTransferRequest)
		v, err := svc.DoTransfer(req.FromAccount, req.ToAccount, req.Amount)
		if err != nil {
			return submitTransferResponse{v, err.Error()}, nil
		}
		return submitTransferResponse{v, ""}, nil
	}
}
