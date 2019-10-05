package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

type Address struct {
	XMLName            xml.Name
	Id                 string     `xml:"id,omitempty"`
	CustomerId         string     `xml:"customer-id,omitempty"`
	FirstName          string     `xml:"first-name,omitempty"`
	LastName           string     `xml:"last-name,omitempty"`
	Company            string     `xml:"company,omitempty"`
	StreetAddress      string     `xml:"street-address,omitempty"`
	ExtendedAddress    string     `xml:"extended-address,omitempty"`
	Locality           string     `xml:"locality,omitempty"`
	Region             string     `xml:"region,omitempty"`
	PostalCode         string     `xml:"postal-code,omitempty"`
	CountryCodeAlpha2  string     `xml:"country-code-alpha2,omitempty"`
	CountryCodeAlpha3  string     `xml:"country-code-alpha3,omitempty"`
	CountryCodeNumeric string     `xml:"country-code-numeric,omitempty"`
	CountryName        string     `xml:"country-name,omitempty"`
	CreatedAt          *time.Time `xml:"created-at,omitempty"`
	UpdatedAt          *time.Time `xml:"updated-at,omitempty"`
}

type AddressRequest struct {
	XMLName            xml.Name `xml:"address"`
	FirstName          string   `xml:"first-name,omitempty"`
	LastName           string   `xml:"last-name,omitempty"`
	Company            string   `xml:"company,omitempty"`
	StreetAddress      string   `xml:"street-address,omitempty"`
	ExtendedAddress    string   `xml:"extended-address,omitempty"`
	Locality           string   `xml:"locality,omitempty"`
	Region             string   `xml:"region,omitempty"`
	PostalCode         string   `xml:"postal-code,omitempty"`
	CountryCodeAlpha2  string   `xml:"country-code-alpha2,omitempty"`
	CountryCodeAlpha3  string   `xml:"country-code-alpha3,omitempty"`
	CountryCodeNumeric string   `xml:"country-code-numeric,omitempty"`
	CountryName        string   `xml:"country-name,omitempty"`
}

type Addresses struct {
	XMLName string     `xml:"addresses"`
	Address []*Address `xml:"address"`
}

const addressesPath = "addresses"

func (c *APIClient) CreateAddress(ctx context.Context, custID string, request *AddressRequest) (*Address, error) {
	response, err := c.do(ctx, http.MethodPost, customersPath+"/"+custID+"/"+addressesPath, &request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result Address
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) DeleteAddress(ctx context.Context, custID, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, customersPath+"/"+custID+"/"+addressesPath+"/"+id, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}

func (c *APIClient) UpdateAddress(ctx context.Context, custID, id string, request *AddressRequest) (*Address, error) {
	response, err := c.do(ctx, http.MethodPut, customersPath+"/"+custID+"/"+addressesPath+"/"+id, request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Address
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}
