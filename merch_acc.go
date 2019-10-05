package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type MerchantAccount struct {
	XMLName                 string                         `xml:"merchant-account,omitempty"`
	Id                      string                         `xml:"id,omitempty"`
	MasterMerchantAccountId string                         `xml:"master-merchant-account-id,omitempty"`
	TOSAccepted             bool                           `xml:"tos_accepted,omitempty"`
	Individual              *MerchantAccountPerson         `xml:"individual,omitempty"`
	Business                *MerchantAccountBusiness       `xml:"business,omitempty"`
	FundingOptions          *MerchantAccountFundingOptions `xml:"funding,omitempty"`
	Status                  string                         `xml:"status,omitempty"`
}

type MerchantAccountPerson struct {
	FirstName   string   `xml:"first-name,omitempty"`
	LastName    string   `xml:"last-name,omitempty"`
	Email       string   `xml:"email,omitempty"`
	Phone       string   `xml:"phone,omitempty"`
	DateOfBirth string   `xml:"date-of-birth,omitempty"`
	SSN         string   `xml:"ssn,omitempty"`
	Address     *Address `xml:"address,omitempty"`
}

type MerchantAccountBusiness struct {
	LegalName string   `xml:"legal-name,omitempty"`
	DbaName   string   `xml:"dba-name,omitempty"`
	TaxId     string   `xml:"tax-id,omitempty"`
	Address   *Address `xml:"address,omitempty"`
}

type MerchantAccountFundingOptions struct {
	Destination   string `xml:"destination,omitempty"`
	Email         string `xml:"email,omitempty"`
	MobilePhone   string `xml:"mobile-phone,omitempty"`
	AccountNumber string `xml:"account-number,omitempty"`
	RoutingNumber string `xml:"routing-number,omitempty"`
}

const (
	FundingDestBank        = "bank"
	FundingDestMobilePhone = "mobile_phone"
	FundingDestEmail       = "email"
)

const (
	merchantAccounts = "merchant_accounts"
	createViaAPI     = "create_via_api"
	updateViaAPI     = "update_via_api"
)

func (c *APIClient) CreateMerchantAccount(ctx context.Context, account *MerchantAccount) (*MerchantAccount, error) {
	cleanAddress(account)
	response, err := c.do(ctx, http.MethodPost, merchantAccounts+"/"+createViaAPI, account)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result MerchantAccount
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindMerchantAccount(ctx context.Context, id string) (*MerchantAccount, error) {
	response, err := c.do(ctx, http.MethodGet, merchantAccounts+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result MerchantAccount
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) UpdateMerchantAccount(ctx context.Context, account *MerchantAccount) (*MerchantAccount, error) {
	cleanAddress(account)
	response, err := c.do(ctx, http.MethodPut, merchantAccounts+"/"+account.Id+"/"+updateViaAPI, account)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result MerchantAccount
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func cleanAddress(account *MerchantAccount) {
	var address *Address
	if account.Individual != nil && account.Individual.Address != nil {
		address = account.Individual.Address
	} else if account.Business != nil && account.Business.Address != nil {
		address = account.Business.Address
	}
	if address != nil && len(address.ExtendedAddress) > 0 {
		address.StreetAddress += " " + address.ExtendedAddress
		address.ExtendedAddress = ""
	}
}
