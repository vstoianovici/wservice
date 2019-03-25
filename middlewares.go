package wservice

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Middlewares is responsibile for creating the wallet service's endpoints that will be wrapped in addtional functionality with go-kit (HTTP transport, circuit-breaking, ratel imitting middlewares)

// For each method, we define request struct that is needed by the MakeTransfersEndpoint enpoint constructor (biolerplate)
type transfersRequest struct {
	S string `json:"s"`
}

// For each method, we define response struct that is needed by the MakeTransfersEndpoint enpoint constructor (biolerplate)
type transfersResponse struct {
	V   []string `json:"v"`
	Err string   `json:"err,omitempty"` // errors don't define JSON marshaling
}

// For each method, we define request struct that is needed by the MakeAccountsEndpoint enpoint constructor (biolerplate)
type accountsRequest struct {
	S string `json:"s"`
}

// For each method, we define response struct that is needed by the MakeAccountsEndpoint enpoint constructor (biolerplate)
type accountsResponse struct {
	V   []string `json:"v"`
	Err string   `json:"err,omitempty"` // errors don't define JSON marshaling
}

// For each method, we define request struct that is needed by the MakeSubmitTransferEndpoint enpoint constructor (biolerplate)
type submitTransferRequest struct {
	FromAccount string `json:"from"`
	ToAccount   string `json:"to"`
	Amount      string `json:"amount"`
}

// For each method, we define response struct that is needed by the MakeSubmitTransferEndpoint enpoint constructor (biolerplate)
type submitTransferResponse struct {
	V   string `json:"result"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

// MakeTransfersEndpoint is an endpoint constructor that takes a service and constructs individual endpoints for the method GetTable method
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

// MakeAccountsEndpoint is an endpoint constructor that takes a service and constructs individual endpoints for the method GetTable method
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

// MakeSubmitTransferEndpoint is an endpoint constructor that takes a service and constructs individual endpoints for the method GetTable method
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
