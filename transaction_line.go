package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type LineItemKind string

const (
	LineItemDebitKind  LineItemKind = "debit"
	LineItemCreditKind LineItemKind = "credit"
	lineItemsPath                   = "line_items"
)

type LineItem struct {
	Kind           LineItemKind `xml:"kind"`
	Name           string       `xml:"name"`
	Description    string       `xml:"description"`
	UnitOfMeasure  string       `xml:"unit-of-measure"`
	ProductCode    string       `xml:"product-code"`
	CommodityCode  string       `xml:"commodity-code"`
	URL            string       `xml:"url"`
	Quantity       *Decimal     `xml:"quantity"`
	UnitAmount     *Decimal     `xml:"unit-amount"`
	UnitTaxAmount  *Decimal     `xml:"unit-tax-amount"`
	TotalAmount    *Decimal     `xml:"total-amount"`
	TaxAmount      *Decimal     `xml:"tax-amount"`
	DiscountAmount *Decimal     `xml:"discount-amount"`
}

type LineItemRequest struct {
	Kind           LineItemKind `xml:"kind"`
	Name           string       `xml:"name"`
	Description    string       `xml:"description,omitempty"`
	UnitOfMeasure  string       `xml:"unit-of-measure,omitempty"`
	ProductCode    string       `xml:"product-code,omitempty"`
	CommodityCode  string       `xml:"commodity-code,omitempty"`
	URL            string       `xml:"url,omitempty"`
	Quantity       *Decimal     `xml:"quantity"`
	UnitAmount     *Decimal     `xml:"unit-amount"`
	UnitTaxAmount  *Decimal     `xml:"unit-tax-amount,omitempty"`
	TotalAmount    *Decimal     `xml:"total-amount"`
	TaxAmount      *Decimal     `xml:"tax-amount,omitempty"`
	DiscountAmount *Decimal     `xml:"discount-amount,omitempty"`
}

type LineItems []*LineItem

func (l *LineItems) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	result := struct {
		LineItems []*LineItem `xml:"line-item"`
	}{}

	err := d.DecodeElement(&result, &start)
	if err != nil {
		return err
	}

	*l = LineItems(result.LineItems)

	return nil
}

type LineItemRequests []*LineItemRequest

func (r LineItemRequests) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(r) == 0 {
		return nil
	}

	result := struct {
		XMLName   string             `xml:"line-items"`
		Type      string             `xml:"type,attr"`
		LineItems []*LineItemRequest `xml:"item"`
	}{
		Type:      "array",
		LineItems: []*LineItemRequest(r),
	}

	return e.EncodeElement(result, start)
}

func (c *APIClient) FindTransactionLineItem(ctx context.Context, txId string) (LineItems, error) {
	response, err := c.do(ctx, http.MethodGet, transactionsPath+"/"+txId+"/"+lineItemsPath, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result LineItems
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, &invalidResponseError{response}
}
