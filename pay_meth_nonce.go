package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type PaymentMethodNonce struct {
	Type             string                     `xml:"type"`
	Nonce            string                     `xml:"nonce"`
	Details          *PaymentMethodNonceDetails `xml:"details"`
	ThreeDSecureInfo *ThreeDSecureInfo          `xml:"three-d-secure-info"`
}

type PaymentMethodNonceDetails struct {
	CardType string `xml:"card-type"`
	Last2    string `xml:"last-two"`
}

func (c *APIClient) FindPaymentMethodNonce(ctx context.Context, nonce string) (*PaymentMethodNonce, error) {
	response, err := c.call(ctx, http.MethodGet, "/payment_method_nonces/"+nonce, nil, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var n PaymentMethodNonce
		if err := xml.Unmarshal(response.Body, &n); err != nil {
			return nil, err
		}
		return &n, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) CreatePaymentMethodNonce(ctx context.Context, token string) (*PaymentMethodNonce, error) {
	response, err := c.call(ctx, http.MethodPost, paymentMethodsPath+"/"+token+"/nonces", nil, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var n PaymentMethodNonce
		if err := xml.Unmarshal(response.Body, &n); err != nil {
			return nil, err
		}
		return &n, nil
	}
	return nil, &invalidResponseError{response}
}
