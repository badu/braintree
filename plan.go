package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

type Plan struct {
	XMLName               string       `xml:"plan"`
	Id                    string       `xml:"id"`
	MerchantId            string       `xml:"merchant-id"`
	BillingDayOfMonth     *int         `xml:"billing-day-of-month"`
	BillingFrequency      *int         `xml:"billing-frequency"`
	CurrencyISOCode       string       `xml:"currency-iso-code"`
	Description           string       `xml:"description"`
	Name                  string       `xml:"name"`
	NumberOfBillingCycles *int         `xml:"number-of-billing-cycles"`
	Price                 *Decimal     `xml:"price"`
	TrialDuration         *int         `xml:"trial-duration"`
	TrialDurationUnit     string       `xml:"trial-duration-unit"`
	TrialPeriod           bool         `xml:"trial-period"`
	CreatedAt             *time.Time   `xml:"created-at"`
	UpdatedAt             *time.Time   `xml:"updated-at"`
	AddOns                AddOnList    `xml:"add-ons"`
	Discounts             DiscountList `xml:"discounts"`
}

type Plans struct {
	XMLName string  `xml:"plans"`
	Plan    []*Plan `xml:"plan"`
}

func (c *APIClient) ListPlans(ctx context.Context) ([]*Plan, error) {
	response, err := c.do(ctx, http.MethodGet, "plans", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var b Plans
		if err := xml.Unmarshal(response.Body, &b); err != nil {
			return nil, err
		}
		return b.Plan, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindPlan(ctx context.Context, id string) (*Plan, error) {
	plans, err := c.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Id == id {
			return p, nil
		}
	}
	return nil, nil
}
