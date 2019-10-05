package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

func (c *APIClient) SandboxSettle(ctx context.Context, transactionID string) (*Tx, error) {
	return c.setStatus(ctx, transactionID, "settle")
}

func (c *APIClient) SandboxSettlementConfirm(ctx context.Context, transactionID string) (*Tx, error) {
	return c.setStatus(ctx, transactionID, "settlement_confirm")
}

func (c *APIClient) SandboxSettlementDecline(ctx context.Context, transactionID string) (*Tx, error) {
	return c.setStatus(ctx, transactionID, "settlement_decline")
}

func (c *APIClient) SandboxSettlementPending(ctx context.Context, transactionID string) (*Tx, error) {
	return c.setStatus(ctx, transactionID, StatusSettlementPending)
}

func (c *APIClient) setStatus(ctx context.Context, transactionID string, status Status) (*Tx, error) {
	if c.Key.URL != ProductionURL {
		response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+transactionID+"/"+string(status), nil)
		if err != nil {
			return nil, err
		}
		switch response.StatusCode {
		case 200:
			var result Tx
			if err := xml.Unmarshal(response.Body, &result); err != nil {
				return nil, err
			}
			return &result, nil
		}
		return nil, &invalidResponseError{response}
	} else {
		return nil, &testOperationPerformedInProductionError{}
	}
}
