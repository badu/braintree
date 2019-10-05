package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type Record struct {
	Count             int      `xml:"count"`
	XMLName           string   `xml:"record"`
	CardType          string   `xml:"card-type"`
	MerchantAccountId string   `xml:"merchant-account-id"`
	Kind              string   `xml:"kind"`
	AmountSettled     *Decimal `xml:"amount-settled"`
}

type XMLRecords struct {
	XMLName string   `xml:"records"`
	Type    []Record `xml:"record"`
}

type SettlementBatchSummary struct {
	XMLName string     `xml:"settlement-batch-summary"`
	Records XMLRecords `xml:"records"`
}

type Settlement struct {
	XMLName string `xml:"settlement_batch_summary"`
	Date    string `xml:"settlement_date"`
}

func (c *APIClient) GenerateSettlement(ctx context.Context, s *Settlement) (*SettlementBatchSummary, error) {
	response, err := c.do(ctx, http.MethodPost, "settlement_batch_summary", s)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result SettlementBatchSummary
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}
