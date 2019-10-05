package braintree

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Response struct {
	*http.Response
	Body []byte
}

func paymentMethod(response *Response) (*PaymentMethod, error) {
	var result struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(response.Body, &result); err != nil {
		return nil, err
	}

	switch result.XMLName.Local {
	case "credit-card":
		var result CreditCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.ToPaymentMethod(), nil
	case "paypal-account":
		var b PayPalAccount
		if err := xml.Unmarshal(response.Body, &b); err != nil {
			return nil, err
		}
		return b.ToPaymentMethod(), nil
	case "venmo-account":
		var result VenmoAccount
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.ToPaymentMethod(), nil
	case "android-pay-card":
		var result AndroidPayCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.ToPaymentMethod(), nil
	case "apple-pay-card":
		var result ApplePayCard
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return result.ToPaymentMethod(), nil
	default:
		return nil, fmt.Errorf("unrecognized payment method %#v", result.XMLName.Local)
	}
}

func (r *Response) unpackBody() error {
	if len(r.Body) == 0 {
		reader := r.Response.Body
		contentEncoding := strings.ToLower(r.Response.Header.Get(HdrContentEncoding))
		if contentEncoding == "gzip" {
			gzipReader, err := gzip.NewReader(reader)
			if err != nil {
				return err
			}
			reader = gzipReader
		}

		defer func() { _ = r.Response.Body.Close() }()

		buf, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		strippedBuf, err := StripNilElements(buf)
		if err == nil {
			r.Body = strippedBuf
		} else {
			r.Body = buf
		}
	}
	return nil
}

func (r *Response) apiError() error {
	var b APIError
	err := xml.Unmarshal(r.Body, &b)
	if err == nil && b.ErrorMessage != "" {
		b.statusCode = r.StatusCode
		return &b
	}
	if r.StatusCode > 299 {
		return httpError(r.StatusCode)
	}
	return nil
}

type APIError struct {
	statusCode int
	errors     ValidationErrors

	ErrorMessage    string
	MerchantAccount *MerchantAccount
	Transaction     *Tx
}

func (e *APIError) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var x struct {
		Errors          ValidationErrors `xml:"errors"`
		ErrorMessage    string           `xml:"message"`
		MerchantAccount *MerchantAccount `xml:"merchant-account"`
		Transaction     *Tx              `xml:"transaction"`
	}
	err := d.DecodeElement(&x, &start)
	if err != nil {
		return err
	}
	e.errors = x.Errors
	e.ErrorMessage = x.ErrorMessage
	e.MerchantAccount = x.MerchantAccount
	e.Transaction = x.Transaction
	return nil
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

func (e *APIError) StatusCode() int {
	return e.statusCode
}

func (e *APIError) All() []ValidationError {
	return e.errors.AllDeep()
}

func (e *APIError) For(name string) *ValidationErrors {
	return e.errors.For(name)
}

type ValidationErrors struct {
	Object           string
	ValidationErrors []ValidationError
	Children         map[string]*ValidationErrors
}

func (r *ValidationErrors) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		t, err := d.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		r.Object = ErrorNameKebabToCamel(start.Name.Local)

		if subStart, ok := t.(xml.StartElement); ok {
			switch subStart.Name.Local {
			case "errors":
				errorList := struct {
					ErrorList []ValidationError `xml:"error"`
				}{}
				err := d.DecodeElement(&errorList, &subStart)
				if err != nil {
					return err
				}
				r.ValidationErrors = errorList.ErrorList
			default:
				subSectionName := ErrorNameKebabToCamel(subStart.Name.Local)
				subSection := &ValidationErrors{}
				err := d.DecodeElement(subSection, &subStart)
				if err != nil {
					return err
				}
				if r.Children == nil {
					r.Children = map[string]*ValidationErrors{}
				}
				r.Children[subSectionName] = subSection
			}
		}
	}
	return nil
}

func (r *ValidationErrors) All() []ValidationError {
	if r == nil {
		return nil
	}
	return r.ValidationErrors
}

func (r *ValidationErrors) AllDeep() []ValidationError {
	if r == nil {
		return nil
	}
	errorList := append([]ValidationError{}, r.All()...)
	for _, sub := range r.Children {
		errorList = append(errorList, sub.AllDeep()...)
	}
	return errorList
}

func (r *ValidationErrors) For(name string) *ValidationErrors {
	if r == nil || r.Children == nil {
		return (*ValidationErrors)(nil)
	}
	return r.Children[name]
}

func (r *ValidationErrors) ForIndex(i int) *ValidationErrors {
	if r == nil || r.Children == nil {
		return (*ValidationErrors)(nil)
	}
	return r.Children["Index"+strconv.Itoa(i)]
}

func (r *ValidationErrors) On(name string) []ValidationError {
	if r == nil {
		return nil
	}
	errors := make([]ValidationError, 0)
	for _, err := range r.ValidationErrors {
		if name == err.Attribute {
			errors = append(errors, err)
		}
	}
	return errors
}

type ValidationError struct {
	Code      string
	Attribute string
	Message   string
}

func (e *ValidationError) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var x struct {
		Code      string `xml:"code"`
		Attribute string `xml:"attribute"`
		Message   string `xml:"message"`
	}
	err := d.DecodeElement(&x, &start)
	if err != nil {
		return err
	}
	e.Code = x.Code
	e.Attribute = ErrorNameSnakeToCamel(x.Attribute)
	e.Message = x.Message
	return nil
}

func ErrorNameSnakeToCamel(src string) string {
	if len(src) == 0 {
		return ""
	}
	camel := bytes.Buffer{}
	capitalizeNext := true
	for i := 0; i < len(src); i++ {
		s := src[i]
		if s == '_' {
			capitalizeNext = true
			continue
		} else if capitalizeNext {
			camel.WriteByte(byte(unicode.ToUpper(rune(s))))
		} else {
			camel.WriteByte(s)
		}
		capitalizeNext = false
	}
	return camel.String()
}

func ErrorNameKebabToCamel(src string) string {
	if len(src) == 0 {
		return ""
	}
	camel := bytes.Buffer{}
	capitalizeNext := true
	for i := 0; i < len(src); i++ {
		k := src[i]
		if k == '-' {
			capitalizeNext = true
			continue
		} else if capitalizeNext {
			camel.WriteByte(byte(unicode.ToUpper(rune(k))))
		} else {
			camel.WriteByte(k)
		}
		capitalizeNext = false
	}
	return camel.String()
}

type IAPIError interface {
	error
	StatusCode() int
}

type httpError int

func (e httpError) StatusCode() int {
	return int(e)
}

func (e httpError) Error() string {
	return fmt.Sprintf("%s (%d)", http.StatusText(e.StatusCode()), e.StatusCode())
}

type invalidResponseError struct {
	resp *Response
}

type InvalidResponseError interface {
	error
	Response() *Response
}

func (e *invalidResponseError) Error() string {
	return fmt.Sprintf("braintree returned invalid response (%d)", e.resp.StatusCode)
}

func (e *invalidResponseError) Response() *Response {
	return e.resp
}

// StripNilElements parses the xml input, removing any elements that
// are decorated with the `nil="true"` attribute returning the XML
// without those elements.
func StripNilElements(data []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	writer := &bytes.Buffer{}
	encoder := xml.NewEncoder(writer)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if IsTokenNil(token) {
			if err := decoder.Skip(); err != nil {
				return nil, err
			}
			continue
		}

		if err := encoder.EncodeToken(token); err != nil {
			return nil, err
		}
	}

	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

func IsTokenNil(src xml.Token) bool {
	if token, ok := src.(xml.StartElement); ok {
		for _, attr := range token.Attr {
			if attr.Name.Space == "" && attr.Name.Local == "nil" && attr.Value == "true" {
				return true
			}
		}
	}
	return false
}

type VenmoAccount struct {
	XMLName           xml.Name       `xml:"venmo-account"`
	CustomerId        string         `xml:"customer-id"`
	Token             string         `xml:"token"`
	Username          string         `xml:"username"`
	VenmoUserID       string         `xml:"venmo-user-id"`
	SourceDescription string         `xml:"source-description"`
	ImageURL          string         `xml:"image-url"`
	CreatedAt         *time.Time     `xml:"created-at"`
	UpdatedAt         *time.Time     `xml:"updated-at"`
	Subscriptions     *Subscriptions `xml:"subscriptions"`
	Default           bool           `xml:"default"`
}

type VenmoAccounts struct {
	Accounts []*VenmoAccount `xml:"venmo-account"`
}

func (v *VenmoAccounts) PaymentMethods() []PaymentMethod {
	if v == nil {
		return nil
	}
	var paymentMethods []PaymentMethod
	for _, account := range v.Accounts {
		paymentMethods = append(paymentMethods, PaymentMethod{
			CustomerId: account.CustomerId,
			Token:      account.Token,
			Default:    account.Default,
			ImageURL:   account.ImageURL,
		})
	}
	return paymentMethods
}

func (v *VenmoAccount) ToPaymentMethod() *PaymentMethod {
	return &PaymentMethod{
		CustomerId: v.CustomerId,
		Token:      v.Token,
		Default:    v.Default,
		ImageURL:   v.ImageURL,
	}
}

func (v *VenmoAccount) AllSubscriptions() []*Subscription {
	if v.Subscriptions == nil {
		return nil
	}
	subscriptions := v.Subscriptions.Subscription
	if len(subscriptions) == 0 {
		return nil
	}
	result := make([]*Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, subscription)
	}
	return result
}
