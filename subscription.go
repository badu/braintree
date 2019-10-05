package braintree

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive       SubscriptionStatus = "Active"
	SubscriptionStatusCanceled     SubscriptionStatus = "Canceled"
	SubscriptionStatusExpired      SubscriptionStatus = "Expired"
	SubscriptionStatusPastDue      SubscriptionStatus = "Past Due"
	SubscriptionStatusPending      SubscriptionStatus = "Pending"
	SubscriptionStatusUnrecognized SubscriptionStatus = "Unrecognized"
)

const (
	Day   = "day"
	Month = "month"
)

type SubscriptionStatusEvent struct {
	Timestamp          time.Time          `xml:"timestamp"`
	Status             SubscriptionStatus `xml:"status"`
	CurrencyISOCode    string             `xml:"currency-iso-code"`
	User               string             `xml:"user"`
	PlanID             string             `xml:"plan-id"`
	SubscriptionSource string             `xml:"subscription-source"`
	Balance            *Decimal           `xml:"balance"`
	Price              *Decimal           `xml:"price"`
}

type Subscription struct {
	XMLName                 string                     `xml:"subscription"`
	Id                      string                     `xml:"id"`
	BillingDayOfMonth       string                     `xml:"billing-day-of-month"`
	BillingPeriodEndDate    string                     `xml:"billing-period-end-date"`
	BillingPeriodStartDate  string                     `xml:"billing-period-start-date"`
	CurrentBillingCycle     string                     `xml:"current-billing-cycle"`
	DaysPastDue             string                     `xml:"days-past-due"`
	FailureCount            string                     `xml:"failure-count"`
	FirstBillingDate        string                     `xml:"first-billing-date"`
	MerchantAccountId       string                     `xml:"merchant-account-id"`
	NextBillingDate         string                     `xml:"next-billing-date"`
	PaidThroughDate         string                     `xml:"paid-through-date"`
	PaymentMethodToken      string                     `xml:"payment-method-token"`
	PlanId                  string                     `xml:"plan-id"`
	TrialDurationUnit       string                     `xml:"trial-duration-unit"`
	TrialDuration           string                     `xml:"trial-duration"`
	Status                  SubscriptionStatus         `xml:"status"`
	NeverExpires            bool                       `xml:"never-expires"`
	TrialPeriod             bool                       `xml:"trial-period"`
	Balance                 *Decimal                   `xml:"balance"`
	NextBillAmount          *Decimal                   `xml:"next-bill-amount"`
	NextBillingPeriodAmount *Decimal                   `xml:"next-billing-period-amount"`
	Price                   *Decimal                   `xml:"price"`
	Transactions            *Transactions              `xml:"transactions"`
	Options                 *SubscriptionOpts          `xml:"options"`
	Descriptor              *Descriptor                `xml:"descriptor"`
	AddOns                  *AddOnList                 `xml:"add-ons"`
	Discounts               *DiscountList              `xml:"discounts"`
	NumberOfBillingCycles   *int                       `xml:"number-of-billing-cycles"`
	CreatedAt               *time.Time                 `xml:"created-at,omitempty"`
	UpdatedAt               *time.Time                 `xml:"updated-at,omitempty"`
	StatusEvents            []*SubscriptionStatusEvent `xml:"status-history>status-event"`
}

type SubscriptionRequest struct {
	XMLName               string                `xml:"subscription"`
	Id                    string                `xml:"id,omitempty"`
	FailureCount          string                `xml:"failure-count,omitempty"`
	FirstBillingDate      string                `xml:"first-billing-date,omitempty"`
	MerchantAccountId     string                `xml:"merchant-account-id,omitempty"`
	PaymentMethodNonce    string                `xml:"paymentMethodNonce,omitempty"`
	PaymentMethodToken    string                `xml:"paymentMethodToken,omitempty"`
	PlanId                string                `xml:"planId,omitempty"`
	TrialDuration         string                `xml:"trial-duration,omitempty"`
	TrialDurationUnit     string                `xml:"trial-duration-unit,omitempty"`
	NeverExpires          *bool                 `xml:"never-expires,omitempty"`
	BillingDayOfMonth     *int                  `xml:"billing-day-of-month,omitempty"`
	NumberOfBillingCycles *int                  `xml:"number-of-billing-cycles,omitempty"`
	Options               *SubscriptionOpts     `xml:"options,omitempty"`
	Price                 *Decimal              `xml:"price,omitempty"`
	TrialPeriod           *bool                 `xml:"trial-period,omitempty"`
	Descriptor            *Descriptor           `xml:"descriptor,omitempty"`
	AddOns                *ModificationsRequest `xml:"add-ons,omitempty"`
	Discounts             *ModificationsRequest `xml:"discounts,omitempty"`
}

type Subscriptions struct {
	Subscription []*Subscription `xml:"subscription"`
}

type SubscriptionOpts struct {
	DoNotInheritAddOnsOrDiscounts        bool `xml:"do-not-inherit-add-ons-or-discounts,omitempty"`
	ProrateCharges                       bool `xml:"prorate-charges,omitempty"`
	ReplaceAllAddOnsAndDiscounts         bool `xml:"replace-all-add-ons-and-discounts,omitempty"`
	RevertSubscriptionOnProrationFailure bool `xml:"revert-subscription-on-proration-failure,omitempty"`
	StartImmediately                     bool `xml:"start-immediately,omitempty"`
}

type SubscriptionTransactionRequest struct {
	SubscriptionID string
	Amount         *Decimal
	Options        *SubscriptionTransactionOptionsRequest
}

func (s *SubscriptionTransactionRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	x := struct {
		XMLName        xml.Name                               `xml:"transaction"`
		Type           string                                 `xml:"type"`
		SubscriptionID string                                 `xml:"subscription-id"`
		Amount         *Decimal                               `xml:"amount,omitempty"`
		Options        *SubscriptionTransactionOptionsRequest `xml:"options,omitempty"`
	}{
		Type:           "sale",
		SubscriptionID: s.SubscriptionID,
		Amount:         s.Amount,
		Options:        s.Options,
	}

	return e.Encode(x)
}

type SubscriptionTransactionOptionsRequest struct {
	SubmitForSettlement bool `xml:"submit-for-settlement"`
}

type SubscriptionSearchResult struct {
	TotalItems        int
	CurrentPageNumber int
	PageSize          int
	TotalIDs          []string
	Subscriptions     []*Subscription
}

const subscriptionsPath = "subscriptions"

func (c *APIClient) CreateSubscription(ctx context.Context, sub *SubscriptionRequest) (*Subscription, error) {
	response, err := c.do(ctx, http.MethodPost, subscriptionsPath, sub)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result Subscription
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) UpdateSubscription(ctx context.Context, subId string, sub *SubscriptionRequest) (*Subscription, error) {
	response, err := c.do(ctx, http.MethodPut, subscriptionsPath+"/"+subId, sub)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Subscription
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindSubscription(ctx context.Context, subId string) (*Subscription, error) {
	response, err := c.do(ctx, http.MethodGet, subscriptionsPath+"/"+subId, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Subscription
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) CancelSubscription(ctx context.Context, subId string) (*Subscription, error) {
	response, err := c.do(ctx, http.MethodPut, subscriptionsPath+"/"+subId+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Subscription
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) RetryCharge(ctx context.Context, txReq *SubscriptionTransactionRequest) error {
	response, err := c.do(ctx, http.MethodPost, transactionsPath, txReq)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 201:
		return nil
	}
	return &invalidResponseError{response}
}

func (c *APIClient) SearchSubscriptions(ctx context.Context, query *Search) (*SearchResult, error) {
	response, err := c.do(ctx, http.MethodPost, subscriptionsPath+"/"+advancedSearchIdsPath, query)
	if err != nil {
		return nil, err
	}

	var searchResult struct {
		PageSize int `xml:"page-size"`
		Ids      struct {
			Item []string `xml:"item"`
		} `xml:"ids"`
	}
	err = xml.Unmarshal(response.Body, &searchResult)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		PageSize:  searchResult.PageSize,
		PageCount: (len(searchResult.Ids.Item) + searchResult.PageSize - 1) / searchResult.PageSize,
		IDs:       searchResult.Ids.Item,
	}, nil
}

func (c *APIClient) SearchSubscription(ctx context.Context, query *Search, searchResult *SearchResult) (*SubscriptionSearchResult, error) {
	if searchResult.Page < 1 || searchResult.Page > searchResult.PageCount {
		return nil, fmt.Errorf("page %d out of bounds, page numbers start at 1 and page count is %d", searchResult.Page, searchResult.PageCount)
	}
	startOffset := (searchResult.Page - 1) * searchResult.PageSize
	endOffset := startOffset + searchResult.PageSize
	if endOffset > len(searchResult.IDs) {
		endOffset = len(searchResult.IDs)
	}

	pageQuery := query.ShallowCopy()
	pageQuery.AddMultiField("ids").Items = searchResult.IDs[startOffset:endOffset]
	subscriptions, err := c.FetchSubscriptions(ctx, pageQuery)

	pageResult := &SubscriptionSearchResult{
		TotalItems:        len(searchResult.IDs),
		TotalIDs:          searchResult.IDs,
		CurrentPageNumber: searchResult.Page,
		PageSize:          searchResult.PageSize,
		Subscriptions:     subscriptions,
	}

	return pageResult, err
}

func (c *APIClient) FetchSubscriptions(ctx context.Context, query *Search) ([]*Subscription, error) {
	response, err := c.do(ctx, http.MethodPost, subscriptionsPath+"/"+advancedSearchPath, query)
	if err != nil {
		return nil, err
	}
	var v struct {
		XMLName       string          `xml:"subscriptions"`
		Subscriptions []*Subscription `xml:"subscription"`
	}
	err = xml.Unmarshal(response.Body, &v)
	if err != nil {
		return nil, err
	}
	return v.Subscriptions, err
}
