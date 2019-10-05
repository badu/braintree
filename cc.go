package braintree

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const monthAndYear = "012006"

func (c *APIClient) CreateCard(ctx context.Context, card *CreditCard) (*CreditCard, error) {
	response, err := c.do(ctx, http.MethodPost, paymentMethodsPath, card)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result CreditCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) UpdateCard(ctx context.Context, card *CreditCard) (*CreditCard, error) {
	response, err := c.do(ctx, http.MethodPut, paymentMethodsPath+"/"+card.Token, card)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result CreditCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindCard(ctx context.Context, token string) (*CreditCard, error) {
	response, err := c.do(ctx, http.MethodGet, paymentMethodsPath+"/"+token, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result CreditCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) DeleteCard(ctx context.Context, card *CreditCard) error {
	resp, err := c.do(ctx, http.MethodDelete, paymentMethodsPath+"/"+card.Token, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}

func (c *APIClient) ExpiringBetween(ctx context.Context, fromDate, toDate time.Time) (*SearchResult, error) {
	qs := url.Values{}
	qs.Set("start", fromDate.UTC().Format(monthAndYear))
	qs.Set("end", toDate.UTC().Format(monthAndYear))
	resp, err := c.do(ctx, http.MethodPost, paymentMethodsPath+"/all/expiring_ids?"+qs.Encode(), nil)
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

func (c *APIClient) ExpiringBetweenPaged(ctx context.Context, fromDate, toDate time.Time, result *SearchResult) (*CreditCardSearchResult, error) {
	if result.Page < 1 || result.Page > result.PageCount {
		return nil, fmt.Errorf("page %d out of bounds, page numbers start at 1 and page count is %d", result.Page, result.PageCount)
	}
	startOffset := (result.Page - 1) * result.PageSize
	endOffset := startOffset + result.PageSize
	if endOffset > len(result.IDs) {
		endOffset = len(result.IDs)
	}

	pageQuery := &Search{}
	pageQuery.AddMultiField("ids").Items = result.IDs[startOffset:endOffset]
	creditCards, err := c.fetchExpiringBetween(ctx, fromDate, toDate, pageQuery)

	pageResult := &CreditCardSearchResult{
		TotalItems:        len(result.IDs),
		TotalIDs:          result.IDs,
		CurrentPageNumber: result.Page,
		PageSize:          result.PageSize,
		CreditCards:       creditCards,
	}

	return pageResult, err
}

func (c *APIClient) fetchExpiringBetween(ctx context.Context, fromDate, toDate time.Time, query *Search) ([]*CreditCard, error) {
	qs := url.Values{}
	qs.Set("start", fromDate.UTC().Format(monthAndYear))
	qs.Set("end", toDate.UTC().Format(monthAndYear))
	resp, err := c.do(ctx, http.MethodPost, paymentMethodsPath+"/all/expiring?"+qs.Encode(), query)
	if err != nil {
		return nil, err
	}

	var v struct {
		CreditCards []*CreditCard `xml:"credit-card"`
	}

	err = xml.Unmarshal(resp.Body, &v)
	if err != nil {
		return nil, err
	}

	return v.CreditCards, nil
}
