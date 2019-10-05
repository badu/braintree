package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type PaymentMethod struct {
	CustomerId string
	Token      string
	Default    bool
	ImageURL   string
}

const paymentMethodsPath = "payment_methods"

type PaymentMethodRequest struct {
	XMLName            xml.Name                     `xml:"payment-method"`
	CustomerId         string                       `xml:"customer-id,omitempty"`
	Token              string                       `xml:"token,omitempty"`
	PaymentMethodNonce string                       `xml:"payment-method-nonce,omitempty"`
	Options            *PaymentMethodRequestOptions `xml:"options,omitempty"`
}

type PaymentMethodRequestOptions struct {
	VerificationMerchantAccountId string `xml:"verification-merchant-account-id,omitempty"`
	MakeDefault                   bool   `xml:"make-default,omitempty"`
	FailOnDuplicatePaymentMethod  bool   `xml:"fail-on-duplicate-payment-method,omitempty"`
	VerifyCard                    *bool  `xml:"verify-card,omitempty"`
}

func (c *APIClient) CreatePayMethod(ctx context.Context, paymentMethodRequest *PaymentMethodRequest) (*PaymentMethod, error) {
	response, err := c.call(ctx, http.MethodPost, paymentMethodsPath, paymentMethodRequest, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		return paymentMethod(response)
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) UpdatePayMethod(ctx context.Context, token string, method *PaymentMethodRequest) (*PaymentMethod, error) {
	response, err := c.call(ctx, http.MethodPut, paymentMethodsPath+"/any/"+token, method, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		return paymentMethod(response)
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindPayMethod(ctx context.Context, token string) (*PaymentMethod, error) {
	response, err := c.call(ctx, http.MethodGet, paymentMethodsPath+"/any/"+token, nil, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		return paymentMethod(response)
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) DeletePayMethod(ctx context.Context, token string) error {
	response, err := c.call(ctx, http.MethodDelete, paymentMethodsPath+"/any/"+token, nil, apiVersion4)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{response}
}
