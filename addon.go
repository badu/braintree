package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
)

type AddOnList struct {
	XMLName string  `xml:"add-ons"`
	AddOns  []AddOn `xml:"add-on"`
}

type AddOn struct {
	XMLName string `xml:"add-on"`
	Modification
}

const addOnsPath = "add_ons"

func (c *APIClient) ListAddons(ctx context.Context) ([]AddOn, error) {
	response, err := c.do(ctx, http.MethodGet, addOnsPath, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result AddOnList
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.AddOns, nil
	}
	return nil, &invalidResponseError{response}
}
