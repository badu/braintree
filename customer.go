package braintree

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const customersPath = "customers"
const advancedSearchPath = "advanced_search"
const advancedSearchIdsPath = "advanced_search_ids"

type AndroidPayCard struct {
	XMLName             xml.Name       `xml:"android-pay-card"`
	Token               string         `xml:"token"`
	CardType            string         `xml:"-"`
	Last4               string         `xml:"-"`
	SourceCardType      string         `xml:"source-card-type"`
	SourceCardLast4     string         `xml:"source-card-last-4"`
	SourceDescription   string         `xml:"source-description"`
	VirtualCardType     string         `xml:"virtual-card-type"`
	VirtualCardLast4    string         `xml:"virtual-card-last-4"`
	ExpirationMonth     string         `xml:"expiration-month"`
	ExpirationYear      string         `xml:"expiration-year"`
	BIN                 string         `xml:"bin"`
	GoogleTransactionID string         `xml:"google-transaction-id"`
	ImageURL            string         `xml:"image-url"`
	Default             bool           `xml:"default"`
	CustomerId          string         `xml:"customer-id"`
	CreatedAt           *time.Time     `xml:"created-at"`
	UpdatedAt           *time.Time     `xml:"updated-at"`
	Subscriptions       *Subscriptions `xml:"subscriptions"`
}

type AndroidPayCards struct {
	AndroidPayCard []*AndroidPayCard `xml:"android-pay-card"`
}

func (a *AndroidPayCards) PaymentMethods() []PaymentMethod {
	if a == nil {
		return nil
	}
	var paymentMethods []PaymentMethod
	for _, ac := range a.AndroidPayCard {
		paymentMethods = append(paymentMethods, PaymentMethod{
			CustomerId: ac.CustomerId,
			Token:      ac.Token,
			Default:    ac.Default,
			ImageURL:   ac.ImageURL,
		})
	}
	return paymentMethods
}

func (a *AndroidPayCard) AllSubscriptions() []*Subscription {
	if a.Subscriptions != nil {
		subs := a.Subscriptions.Subscription
		if len(subs) > 0 {
			a := make([]*Subscription, 0, len(subs))
			for _, s := range subs {
				a = append(a, s)
			}
			return a
		}
	}
	return nil
}

func (a *AndroidPayCard) ToPaymentMethod() *PaymentMethod {
	return &PaymentMethod{
		CustomerId: a.CustomerId,
		Token:      a.Token,
		Default:    a.Default,
		ImageURL:   a.ImageURL,
	}
}

func (a *AndroidPayCard) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	type typeWithNoFunctions AndroidPayCard
	if err := decoder.DecodeElement((*typeWithNoFunctions)(a), &start); err != nil {
		return err
	}
	a.CardType = a.VirtualCardType
	a.Last4 = a.VirtualCardLast4
	return nil
}

type ApplePayCard struct {
	XMLName               xml.Name       `xml:"apple-pay-card"`
	Token                 string         `xml:"token"`
	ImageURL              string         `xml:"image-url"`
	CardType              string         `xml:"card-type"`
	PaymentInstrumentName string         `xml:"payment-instrument-name"`
	SourceDescription     string         `xml:"source-description"`
	BIN                   string         `xml:"bin"`
	Last4                 string         `xml:"last-4"`
	ExpirationMonth       string         `xml:"expiration-month"`
	ExpirationYear        string         `xml:"expiration-year"`
	Expired               bool           `xml:"expired"`
	Default               bool           `xml:"default"`
	CustomerId            string         `xml:"customer-id"`
	CreatedAt             *time.Time     `xml:"created-at"`
	UpdatedAt             *time.Time     `xml:"updated-at"`
	Subscriptions         *Subscriptions `xml:"subscriptions"`
}

type ApplePayCards struct {
	Cards []*ApplePayCard `xml:"apple-pay-card"`
}

func (a *ApplePayCards) PaymentMethods() []PaymentMethod {
	if a == nil {
		return nil
	}
	var paymentMethods []PaymentMethod
	for _, a := range a.Cards {
		paymentMethods = append(paymentMethods, PaymentMethod{
			CustomerId: a.CustomerId,
			Token:      a.Token,
			Default:    a.Default,
			ImageURL:   a.ImageURL,
		})
	}
	return paymentMethods
}

func (a *ApplePayCard) ToPaymentMethod() *PaymentMethod {
	return &PaymentMethod{
		CustomerId: a.CustomerId,
		Token:      a.Token,
		Default:    a.Default,
		ImageURL:   a.ImageURL,
	}
}

func (a *ApplePayCard) AllSubscriptions() []*Subscription {
	if a.Subscriptions != nil {
		subs := a.Subscriptions.Subscription
		if len(subs) > 0 {
			a := make([]*Subscription, 0, len(subs))
			for _, s := range subs {
				a = append(a, s)
			}
			return a
		}
	}
	return nil
}

type Customer struct {
	XMLName            string           `xml:"customer"`
	Id                 string           `xml:"id"`
	FirstName          string           `xml:"first-name"`
	LastName           string           `xml:"last-name"`
	Company            string           `xml:"company"`
	Email              string           `xml:"email"`
	Phone              string           `xml:"phone"`
	Fax                string           `xml:"fax"`
	Website            string           `xml:"website"`
	PaymentMethodNonce string           `xml:"payment-method-nonce"`
	CustomFields       CustomFields     `xml:"custom-fields"`
	CreditCard         *CreditCard      `xml:"credit-card"`
	CreditCards        *CreditCards     `xml:"credit-cards"`
	PayPalAccounts     *PayPalAccounts  `xml:"paypal-accounts"`
	VenmoAccounts      *VenmoAccounts   `xml:"venmo-accounts"`
	AndroidPayCards    *AndroidPayCards `xml:"android-pay-cards"`
	ApplePayCards      *ApplePayCards   `xml:"apple-pay-cards"`
	Addresses          *Addresses       `xml:"addresses"`
	CreatedAt          *time.Time       `xml:"created-at"`
	UpdatedAt          *time.Time       `xml:"updated-at"`
}

func (c *Customer) PaymentMethods() []PaymentMethod {
	var paymentMethods []PaymentMethod
	paymentMethods = append(paymentMethods, c.CreditCards.PaymentMethods()...)
	paymentMethods = append(paymentMethods, c.PayPalAccounts.PaymentMethods()...)
	paymentMethods = append(paymentMethods, c.VenmoAccounts.PaymentMethods()...)
	paymentMethods = append(paymentMethods, c.AndroidPayCards.PaymentMethods()...)
	paymentMethods = append(paymentMethods, c.ApplePayCards.PaymentMethods()...)
	return paymentMethods
}

func (c *Customer) DefaultCreditCard() *CreditCard {
	for _, card := range c.CreditCards.CreditCard {
		if card.Default {
			return card
		}
	}
	return nil
}

func (c *Customer) DefaultPaymentMethod() *PaymentMethod {
	for _, pm := range c.PaymentMethods() {
		if pm.Default {
			return &pm
		}
	}
	return nil
}

type CreditCards struct {
	CreditCard []*CreditCard `xml:"credit-card"`
}

func (c *CreditCards) PaymentMethods() []PaymentMethod {
	if c == nil {
		return nil
	}
	var paymentMethods []PaymentMethod
	for _, cc := range c.CreditCard {
		paymentMethods = append(paymentMethods, PaymentMethod{
			CustomerId: cc.CustomerId,
			Token:      cc.Token,
			Default:    cc.Default,
			ImageURL:   cc.ImageURL,
		})
	}
	return paymentMethods
}

type CreditCardOptions struct {
	VenmoSDKSession               string `xml:"venmo-sdk-session,omitempty"`
	VerificationMerchantAccountId string `xml:"verification-merchant-account-id,omitempty"`
	UpdateExistingToken           string `xml:"update-existing-token,omitempty"`
	MakeDefault                   bool   `xml:"make-default,omitempty"`
	FailOnDuplicatePaymentMethod  bool   `xml:"fail-on-duplicate-payment-method,omitempty"`
	VerifyCard                    *bool  `xml:"verify-card,omitempty"`
}

func (c *CreditCard) AllSubscriptions() []*Subscription {
	if c.Subscriptions != nil {
		subs := c.Subscriptions.Subscription
		if len(subs) > 0 {
			a := make([]*Subscription, 0, len(subs))
			for _, s := range subs {
				a = append(a, s)
			}
			return a
		}
	}
	return nil
}

func (cc *CreditCard) ToPaymentMethod() *PaymentMethod {
	return &PaymentMethod{
		CustomerId: cc.CustomerId,
		Token:      cc.Token,
		Default:    cc.Default,
		ImageURL:   cc.ImageURL,
	}
}

type CreditCardSearchResult struct {
	TotalItems        int
	CurrentPageNumber int
	PageSize          int
	TotalIDs          []string
	CreditCards       []*CreditCard
}

type CreditCard struct {
	XMLName                   xml.Name           `xml:"credit-card"`
	CustomerId                string             `xml:"customer-id,omitempty"`
	Token                     string             `xml:"token,omitempty"`
	PaymentMethodNonce        string             `xml:"payment-method-nonce,omitempty"`
	Number                    string             `xml:"number,omitempty"`
	ExpirationDate            string             `xml:"expiration-date,omitempty"`
	ExpirationMonth           string             `xml:"expiration-month,omitempty"`
	ExpirationYear            string             `xml:"expiration-year,omitempty"`
	CVV                       string             `xml:"cvv,omitempty"`
	VenmoSDKPaymentMethodCode string             `xml:"venmo-sdk-payment-method-code,omitempty"`
	Last4                     string             `xml:"last-4,omitempty"`
	Commercial                string             `xml:"commercial,omitempty"`
	Debit                     string             `xml:"debit,omitempty"`
	DurbinRegulated           string             `xml:"durbin-regulated,omitempty"`
	Healthcare                string             `xml:"healthcare,omitempty"`
	Payroll                   string             `xml:"payroll,omitempty"`
	Prepaid                   string             `xml:"prepaid,omitempty"`
	CountryOfIssuance         string             `xml:"country-of-issuance,omitempty"`
	IssuingBank               string             `xml:"issuing-bank,omitempty"`
	UniqueNumberIdentifier    string             `xml:"unique-number-identifier,omitempty"`
	Bin                       string             `xml:"bin,omitempty"`
	CardType                  string             `xml:"card-type,omitempty"`
	CardholderName            string             `xml:"cardholder-name,omitempty"`
	CustomerLocation          string             `xml:"customer-location,omitempty"`
	ImageURL                  string             `xml:"image-url,omitempty"`
	ProductID                 string             `xml:"product-id,omitempty"`
	VenmoSDK                  bool               `xml:"venmo-sdk,omitempty"`
	Default                   bool               `xml:"default,omitempty"`
	Expired                   bool               `xml:"expired,omitempty"`
	Options                   *CreditCardOptions `xml:"options,omitempty"`
	UpdatedAt                 *time.Time         `xml:"updated-at,omitempty"`
	CreatedAt                 *time.Time         `xml:"created-at,omitempty"`
	BillingAddress            *Address           `xml:"billing-address,omitempty"`
	Subscriptions             *Subscriptions     `xml:"subscriptions,omitempty"`
}

type xmlField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type CustomFields map[string]string

func (c CustomFields) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if len(c) == 0 {
		return nil
	}

	err := encoder.EncodeToken(start)
	if err != nil {
		return err
	}

	var nameMarshalReplacer = strings.NewReplacer("_", "-")
	for k, v := range c {
		tag := nameMarshalReplacer.Replace(k)
		err := encoder.Encode(xmlField{
			XMLName: xml.Name{Local: tag},
			Value:   v,
		})
		if err != nil {
			return err
		}
	}

	return encoder.EncodeToken(start.End())
}

func (c *CustomFields) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	*c = CustomFields{}

	var nameUnmarshalReplacer = strings.NewReplacer("-", "_")
	for {
		var cf xmlField

		err := decoder.Decode(&cf)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		name := nameUnmarshalReplacer.Replace(cf.XMLName.Local)
		(*c)[name] = cf.Value
	}
	return nil
}

type CustomerRequest struct {
	XMLName            string       `xml:"customer"`
	ID                 string       `xml:"id,omitempty"`
	FirstName          string       `xml:"first-name,omitempty"`
	LastName           string       `xml:"last-name,omitempty"`
	Company            string       `xml:"company,omitempty"`
	Email              string       `xml:"email,omitempty"`
	Phone              string       `xml:"phone,omitempty"`
	Fax                string       `xml:"fax,omitempty"`
	Website            string       `xml:"website,omitempty"`
	PaymentMethodNonce string       `xml:"payment-method-nonce,omitempty"`
	CustomFields       CustomFields `xml:"custom-fields,omitempty"`
	CreditCard         *CreditCard  `xml:"credit-card,omitempty"`
}

type CustomerSearchResult struct {
	TotalItems        int
	CurrentPageNumber int
	PageSize          int
	TotalIDs          []string
	Customers         []*Customer
}

func (c *APIClient) CreateCustomer(ctx context.Context, request *CustomerRequest) (*Customer, error) {
	response, err := c.do(ctx, http.MethodPost, customersPath, request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result Customer
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) UpdateCustomer(ctx context.Context, request *CustomerRequest) (*Customer, error) {
	response, err := c.do(ctx, http.MethodPut, customersPath+"/"+request.ID, request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Customer
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindCustomer(ctx context.Context, id string) (*Customer, error) {
	response, err := c.do(ctx, http.MethodGet, customersPath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Customer
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) SearchCustomersByIDs(ctx context.Context, query *Search) (*SearchResult, error) {
	resp, err := c.do(ctx, http.MethodPost, customersPath+"/"+advancedSearchIdsPath, query)
	if err != nil {
		return nil, err
	}

	var searchResult struct {
		PageSize int `xml:"page-size"`
		Ids      struct {
			Item []string `xml:"item"`
		} `xml:"ids"`
	}
	err = xml.Unmarshal(resp.Body, &searchResult)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		PageSize:  searchResult.PageSize,
		PageCount: (len(searchResult.Ids.Item) + searchResult.PageSize - 1) / searchResult.PageSize,
		IDs:       searchResult.Ids.Item,
	}, nil
}

func (c *APIClient) SearchCustomer(ctx context.Context, query *Search, result *SearchResult) (*CustomerSearchResult, error) {
	if result.Page < 1 || result.Page > result.PageCount {
		return nil, fmt.Errorf("page %d out of bounds, page numbers start at 1 and page count is %d", result.Page, result.PageCount)
	}
	startOffset := (result.Page - 1) * result.PageSize
	endOffset := startOffset + result.PageSize
	if endOffset > len(result.IDs) {
		endOffset = len(result.IDs)
	}

	pageQuery := query.ShallowCopy()
	pageQuery.AddMultiField("ids").Items = result.IDs[startOffset:endOffset]
	customers, err := c.fetchCustomers(ctx, pageQuery)

	pageResult := &CustomerSearchResult{
		TotalItems:        len(result.IDs),
		TotalIDs:          result.IDs,
		CurrentPageNumber: result.Page,
		PageSize:          result.PageSize,
		Customers:         customers,
	}

	return pageResult, err
}

func (c *APIClient) fetchCustomers(ctx context.Context, query *Search) ([]*Customer, error) {
	resp, err := c.do(ctx, http.MethodPost, customersPath+"/"+advancedSearchPath, query)
	if err != nil {
		return nil, err
	}
	var v struct {
		XMLName   string      `xml:"customers"`
		Customers []*Customer `xml:"customer"`
	}
	err = xml.Unmarshal(resp.Body, &v)
	if err != nil {
		return nil, err
	}
	return v.Customers, err
}

func (c *APIClient) DeleteCustomer(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, customersPath+"/"+id, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}
