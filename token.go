package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

const clientTokenVersion = 2

type TokenRequest struct {
	XMLName           string   `xml:"client-token"`
	CustomerID        string   `xml:"customer-id,omitempty"`
	MerchantAccountID string   `xml:"merchant-account-id,omitempty"`
	Options           *Options `xml:"options,omitempty"`
	Version           int      `xml:"version"`
}

type clientToken struct {
	ClientToken string `xml:"value"`
}

type Options struct {
	FailOnDuplicatePaymentMethod bool  `xml:"fail-on-duplicate-payment-method,omitempty"`
	MakeDefault                  bool  `xml:"make-default,omitempty"`
	VerifyCard                   *bool `xml:"verify-card,omitempty"`
}

const clientTokenPath = "client_token"

func (c *APIClient) GenerateToken(ctx context.Context) (string, error) {
	return c.generate(ctx, &TokenRequest{
		Version: clientTokenVersion,
	})
}

func (c *APIClient) GenerateWithCustomer(ctx context.Context, custID string) (string, error) {
	return c.generate(ctx, &TokenRequest{
		Version:    clientTokenVersion,
		CustomerID: custID,
	})
}

func (c *APIClient) GenerateWithRequest(ctx context.Context, request *TokenRequest) (string, error) {
	if request == nil {
		request = &TokenRequest{}
	}
	if request.Version == 0 {
		request.Version = clientTokenVersion
	}
	return c.generate(ctx, request)
}

func (c *APIClient) generate(ctx context.Context, request *TokenRequest) (string, error) {
	resp, err := c.do(ctx, http.MethodPost, clientTokenPath, request)
	if err != nil {
		return "", err
	}
	switch resp.StatusCode {
	case 201:
		var b clientToken
		if err := xml.Unmarshal(resp.Body, &b); err != nil {
			return "", err
		}
		return b.ClientToken, nil
	}
	return "", &invalidResponseError{resp}
}
