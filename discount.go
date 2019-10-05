package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

const discounts = "discounts"

type DiscountList struct {
	XMLName   string     `xml:"discounts"`
	Discounts []Discount `xml:"discount"`
}

const (
	ModificationKindDiscount = "discount"
	ModificationKindAddOn    = "add_on"
)

type Modification struct {
	Id                    string     `xml:"id"`
	Amount                *Decimal   `xml:"amount"`
	Description           string     `xml:"description"`
	Kind                  string     `xml:"kind"`
	Name                  string     `xml:"name"`
	NeverExpires          bool       `xml:"never-expires"`
	Quantity              int        `xml:"quantity"`
	NumberOfBillingCycles int        `xml:"number-of-billing-cycles"`
	CurrentBillingCycle   int        `xml:"current-billing-cycle"`
	UpdatedAt             *time.Time `xml:"updated_at"`
}

type Discount struct {
	XMLName string `xml:"discount"`
	Modification
}

func (c *APIClient) AllDiscounts(ctx context.Context) ([]Discount, error) {
	response, err := c.do(ctx, http.MethodGet, discounts, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result DiscountList
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.Discounts, nil
	}
	return nil, &invalidResponseError{response}
}
