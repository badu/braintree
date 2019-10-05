package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

type PayPalAccount struct {
	XMLName       xml.Name              `xml:"paypal-account"`
	CustomerId    string                `xml:"customer-id,omitempty"`
	Token         string                `xml:"token,omitempty"`
	Email         string                `xml:"email,omitempty"`
	ImageURL      string                `xml:"image-url,omitempty"`
	Default       bool                  `xml:"default,omitempty"`
	CreatedAt     *time.Time            `xml:"created-at,omitempty"`
	UpdatedAt     *time.Time            `xml:"updated-at,omitempty"`
	Subscriptions *Subscriptions        `xml:"subscriptions,omitempty"`
	Options       *PayPalAccountOptions `xml:"options,omitempty"`
}

type PayPalAccounts struct {
	Accounts []*PayPalAccount `xml:"paypal-account"`
}

func (a *PayPalAccounts) PaymentMethods() []PaymentMethod {
	if a == nil {
		return nil
	}
	var paymentMethods []PaymentMethod
	for _, pp := range a.Accounts {
		paymentMethods = append(paymentMethods, PaymentMethod{
			CustomerId: pp.CustomerId,
			Token:      pp.Token,
			Default:    pp.Default,
			ImageURL:   pp.ImageURL,
		})
	}
	return paymentMethods
}

func (a *PayPalAccount) ToPaymentMethod() *PaymentMethod {
	return &PaymentMethod{
		CustomerId: a.CustomerId,
		Token:      a.Token,
		Default:    a.Default,
		ImageURL:   a.ImageURL,
	}
}

type PayPalAccountOptions struct {
	MakeDefault bool `xml:"make-default,omitempty"`
}

func (a *PayPalAccount) AllSubscriptions() []*Subscription {
	if a.Subscriptions != nil {
		subscriptions := a.Subscriptions.Subscription
		if len(subscriptions) > 0 {
			result := make([]*Subscription, 0, len(subscriptions))
			for _, subscription := range subscriptions {
				result = append(result, subscription)
			}
			return result
		}
	}
	return nil
}

func (c *APIClient) UpdatePaypalAccount(ctx context.Context, paypalAccount *PayPalAccount) (*PayPalAccount, error) {
	response, err := c.call(ctx, http.MethodPut, paymentMethodsPath+"/paypal_account/"+paypalAccount.Token, paypalAccount, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var b PayPalAccount
		if err := xml.Unmarshal(response.Body, &b); err != nil {
			return nil, err
		}
		return &b, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) SearchPaypalAccount(ctx context.Context, token string) (*PayPalAccount, error) {
	response, err := c.call(ctx, http.MethodGet, paymentMethodsPath+"/paypal_account/"+token, nil, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var b PayPalAccount
		if err := xml.Unmarshal(response.Body, &b); err != nil {
			return nil, err
		}
		return &b, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) DeletePaypalAccount(ctx context.Context, paypalAccount *PayPalAccount) error {
	response, err := c.call(ctx, http.MethodDelete, paymentMethodsPath+"/paypal_account/"+paypalAccount.Token, nil, apiVersion4)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{response}
}
