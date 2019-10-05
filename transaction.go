package braintree

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const transactionsPath = "transactions"

type Status string

const (
	StatusAuthorizationExpired   Status = "authorization_expired"
	StatusAuthorizing            Status = "authorizing"
	StatusAuthorized             Status = "authorized"
	StatusGatewayRejected        Status = "gateway_rejected"
	StatusFailed                 Status = "failed"
	StatusProcessorDeclined      Status = "processor_declined"
	StatusSettled                Status = "settled"
	StatusSettlementConfirmed    Status = "settlement_confirmed"
	StatusSettlementDeclined     Status = "settlement_declined"
	StatusSettlementPending      Status = "settlement_pending"
	StatusSettling               Status = "settling"
	StatusSubmittedForSettlement Status = "submitted_for_settlement"
	StatusVoided                 Status = "voided"
	StatusUnknown                Status = "unrecognized"
)

type TxSource string

const (
	RecurringFirst TxSource = "recurring_first"
	Recurring      TxSource = "recurring"
	MOTO           TxSource = "moto"
	Merchant       TxSource = "merchant"
)

type PaymentType string

const (
	AndroidPayCardType   PaymentType = "android_pay_card"
	ApplePayCardType     PaymentType = "apple_pay_card"
	CreditCardType       PaymentType = "credit_card"
	MasterpassCardType   PaymentType = "masterpass_card"
	PaypalAccountType    PaymentType = "paypal_account"
	VenmoAccountType     PaymentType = "venmo_account"
	VisaCheckoutCardType PaymentType = "visa_checkout_card"
)

type RejectionReason string

const (
	ApplicationIncompleteReason RejectionReason = "application_incomplete"
	AVSReason                   RejectionReason = "avs"
	AVSAndCVVReason             RejectionReason = "avs_and_cvv"
	CVVReason                   RejectionReason = "cvv"
	DuplicateReason             RejectionReason = "duplicate"
	FraudReason                 RejectionReason = "fraud"
	ThreeDSecureReason          RejectionReason = "three_d_secure"
	UnknownReason               RejectionReason = "unrecognized"
)

type AndroidPayDetail struct {
	Token               string `xml:"token"`
	CardType            string `xml:"-"`
	Last4               string `xml:"-"`
	SourceCardType      string `xml:"source-card-type"`
	SourceCardLast4     string `xml:"source-card-last-4"`
	SourceDescription   string `xml:"source-description"`
	VirtualCardType     string `xml:"virtual-card-type"`
	VirtualCardLast4    string `xml:"virtual-card-last-4"`
	ExpirationMonth     string `xml:"expiration-month"`
	ExpirationYear      string `xml:"expiration-year"`
	BIN                 string `xml:"bin"`
	GoogleTransactionID string `xml:"google-transaction-id"`
	ImageURL            string `xml:"image-url"`
}

func (a *AndroidPayDetail) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type typeWithNoFunctions AndroidPayDetail
	if err := d.DecodeElement((*typeWithNoFunctions)(a), &start); err != nil {
		return err
	}
	a.CardType = a.VirtualCardType
	a.Last4 = a.VirtualCardLast4
	return nil
}

type ApplePayDetail struct {
	Token                 string `xml:"token"`
	CardType              string `xml:"card-type"`
	PaymentInstrumentName string `xml:"payment-instrument-name"`
	SourceDescription     string `xml:"source-description"`
	CardholderName        string `xml:"cardholder-name"`
	ExpirationMonth       string `xml:"expiration-month"`
	ExpirationYear        string `xml:"expiration-year"`
	Last4                 string `xml:"last-4"`
	BIN                   string `xml:"bin"`
}

type AVSResponseCode string

const (
	AVSResponseCodeMatches       AVSResponseCode = "M" // The postal code or street address provided matches the information on file with the cardholder's bank.
	AVSResponseCodeDoesNotMatch  AVSResponseCode = "N" // The postal code or street address provided does not match the information on file with the cardholder's bank.
	AVSResponseCodeNotVerified   AVSResponseCode = "U" // The card-issuing bank received the postal code or street address but did not verify whether it was correct. This typically happens if the processor declines an authorization before the bank evaluates the postal code.
	AVSResponseCodeNotProvided   AVSResponseCode = "I" // No postal code or street address was provided.
	AVSResponseCodeNotSupported  AVSResponseCode = "S" // AVS information was provided but the card-issuing bank does not participate in address verification. This typically indicates a card-issuing bank outside of the US, Canada, and the UK.
	AVSResponseCodeSystemError   AVSResponseCode = "E" // A system error prevented any verification of street address or postal code.
	AVSResponseCodeNotApplicable AVSResponseCode = "A" // AVS information was provided but this type of transaction does not support address verification.
)

type CVVResponseCode string

const (
	CVVResponseCodeMatches                  CVVResponseCode = "M" // CVVResponseCodeMatches means the CVV provided matches the information on file with the cardholder's bank.
	CVVResponseCodeDoesNotMatch             CVVResponseCode = "N" // CVVResponseCodeDoesNotMatch means the The CVV provided does not match the information on file with the cardholder's bank.
	CVVResponseCodeNotVerified              CVVResponseCode = "U" // CVVResponseCodeNotVerified means the the card-issuing bank received the CVV but did not verify whether it was correct. This typically happens if the processor declines an authorization before the bank evaluates the CVV.
	CVVResponseCodeNotProvided              CVVResponseCode = "I" // CVVResponseCodeNotProvided means the no CVV was provided.
	CVVResponseCodeIssuerDoesNotParticipate CVVResponseCode = "S" // CVVResponseCodeIssuerDoesNotParticipate means the the CVV was provided but the card-issuing bank does not participate in card verification.
	CVVResponseCodeNotApplicable            CVVResponseCode = "A" // CVVResponseCodeNotApplicable means the the CVV was provided but this type of transaction does not support card verification.
)

type Descriptor struct {
	Name  string `xml:"name,omitempty"`
	Phone string `xml:"phone,omitempty"`
	URL   string `xml:"url,omitempty"`
}

type EscrowStatus string

const (
	EscrowHoldPending    EscrowStatus = "hold_pending"
	EscrowHeld           EscrowStatus = "held"
	EscrowReleasePending EscrowStatus = "release_pending"
	EscrowReleased       EscrowStatus = "released"
	EscrowRefunded       EscrowStatus = "refunded"
)

type ThreeDSecureInfo struct {
	Status                 ThreeDSecureStatus   `xml:"status"`
	Enrolled               ThreeDSecureEnrolled `xml:"enrolled"`
	LiabilityShiftPossible bool                 `xml:"liability-shift-possible"`
	LiabilityShifted       bool                 `xml:"liability-shifted"`
}

type ThreeDSecureStatus string

const (
	ThreeDSecureUnsupportedCard                              ThreeDSecureStatus = "unsupported_card"
	ThreeDSecureLookupError                                  ThreeDSecureStatus = "lookup_error"
	ThreeDSecureLookupEnrolled                               ThreeDSecureStatus = "lookup_enrolled"
	ThreeDSecureLookupNotEnrolled                            ThreeDSecureStatus = "lookup_not_enrolled"
	ThreeDSecureAuthenticateSuccessfulIssuerNotParticipating ThreeDSecureStatus = "authenticate_successful_issuer_not_participating"
	ThreeDSecureAuthenticationUnavailable                    ThreeDSecureStatus = "authentication_unavailable"
	ThreeDSecureAuthenticateSignatureVerificationFailed      ThreeDSecureStatus = "authenticate_signature_verification_failed"
	ThreeDSecureAuthenticateSuccessful                       ThreeDSecureStatus = "authenticate_successful"
	ThreeDSecureAuthenticateAttemptSuccessful                ThreeDSecureStatus = "authenticate_attempt_successful"
	ThreeDSecureAuthenticateFailed                           ThreeDSecureStatus = "authenticate_failed"
	ThreeDSecureAuthenticateUnableToAuthenticate             ThreeDSecureStatus = "authenticate_unable_to_authenticate"
	ThreeDSecureAuthenticateError                            ThreeDSecureStatus = "authenticate_error"
)

type ThreeDSecureEnrolled string

const (
	EnrolledYes            ThreeDSecureEnrolled = "Y"
	EnrolledNo             ThreeDSecureEnrolled = "N"
	EnrolledUnavailable    ThreeDSecureEnrolled = "U"
	EnrolledBypass         ThreeDSecureEnrolled = "B"
	EnrolledRequestFailure ThreeDSecureEnrolled = "E"
)

type ResponseCode int

func (rc ResponseCode) Int() int {
	return int(rc)
}

func (rc *ResponseCode) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	code, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}

	*rc = ResponseCode(code)

	return nil
}

func (rc ResponseCode) MarshalText() ([]byte, error) {
	if rc == 0 {
		return nil, nil
	}
	return []byte(strconv.Itoa(int(rc))), nil
}

type ResponseType string

const (
	ResponseTypeApproved     ResponseType = "approved"
	ResponseTypeSoftDeclined ResponseType = "soft_declined"
	ResponseTypeHardDeclined ResponseType = "hard_declined"
)

type DisbursementDetail struct {
	XMLName                        xml.Name `xml:"disbursement-details"`
	DisbursementDate               string   `xml:"disbursement-date"`
	SettlementAmount               *Decimal `xml:"settlement-amount"`
	SettlementCurrencyIsoCode      string   `xml:"settlement-currency-iso-code"`
	SettlementCurrencyExchangeRate *Decimal `xml:"settlement-currency-exchange-rate"`
	FundsHeld                      bool     `xml:"funds-held"`
	Success                        bool     `xml:"success"`
}

type PayPalDetail struct {
	PayerEmail                    string `xml:"payer-email,omitempty"`
	PaymentID                     string `xml:"payment-id,omitempty"`
	AuthorizationID               string `xml:"authorization-id,omitempty"`
	Token                         string `xml:"token,omitempty"`
	ImageURL                      string `xml:"image-url,omitempty"`
	DebugID                       string `xml:"debug-id,omitempty"`
	PayeeEmail                    string `xml:"payee-email,omitempty"`
	CustomField                   string `xml:"custom-field,omitempty"`
	PayerID                       string `xml:"payer-id,omitempty"`
	PayerFirstName                string `xml:"payer-first-name,omitempty"`
	PayerLastName                 string `xml:"payer-last-name,omitempty"`
	PayerStatus                   string `xml:"payer-status,omitempty"`
	SellerProtectionStatus        string `xml:"seller-protection-status,omitempty"`
	RefundID                      string `xml:"refund-id,omitempty"`
	CaptureID                     string `xml:"capture-id,omitempty"`
	TransactionFeeAmount          string `xml:"transaction-fee-amount,omitempty"`
	TransactionFeeCurrencyISOCode string `xml:"transaction-fee-currency-iso-code,omitempty"`
	Description                   string `xml:"description,omitempty"`
}

type VenmoAccountDetail struct {
	Token             string `xml:"token,omitempty"`
	Username          string `xml:"username,omitempty"`
	VenmoUserID       string `xml:"venmo-user-id,omitempty"`
	SourceDescription string `xml:"source-description,omitempty"`
	ImageURL          string `xml:"image-url,omitempty"`
}

type Tx struct {
	ProcessorResponseCode        ResponseCode        `xml:"processor-response-code"`
	ProcessorResponseType        ResponseType        `xml:"processor-response-type"`
	EscrowStatus                 EscrowStatus        `xml:"escrow-status"`
	PaymentInstrumentType        PaymentType         `xml:"payment-instrument-type"`
	AVSErrorResponseCode         AVSResponseCode     `xml:"avs-error-response-code"`
	AVSPostalCodeResponseCode    AVSResponseCode     `xml:"avs-postal-code-response-code"`
	AVSStreetAddressResponseCode AVSResponseCode     `xml:"avs-street-address-response-code"`
	CVVResponseCode              CVVResponseCode     `xml:"cvv-response-code"`
	GatewayRejectionReason       RejectionReason     `xml:"gateway-rejection-reason"`
	CustomFields                 CustomFields        `xml:"custom-fields"`
	ProcessorAuthorizationCode   string              `xml:"processor-authorization-code"`
	SettlementBatchId            string              `xml:"settlement-batch-id"`
	XMLName                      string              `xml:"transaction"`
	Id                           string              `xml:"id"`
	Status                       Status              `xml:"status"`
	Type                         string              `xml:"type"`
	CurrencyISOCode              string              `xml:"currency-iso-code"`
	OrderId                      string              `xml:"order-id"`
	PaymentMethodToken           string              `xml:"payment-method-token"`
	PaymentMethodNonce           string              `xml:"payment-method-nonce"`
	MerchantAccountId            string              `xml:"merchant-account-id"`
	PlanId                       string              `xml:"plan-id"`
	SubscriptionId               string              `xml:"subscription-id"`
	DeviceData                   string              `xml:"device-data"`
	RefundId                     string              `xml:"refund-id"`
	ProcessorResponseText        string              `xml:"processor-response-text"`
	AdditionalProcessorResponse  string              `xml:"additional-processor-response"`
	Channel                      string              `xml:"channel"`
	PurchaseOrderNumber          string              `xml:"purchase-order-number"`
	TaxExempt                    bool                `xml:"tax-exempt"`
	CreatedAt                    *time.Time          `xml:"created-at"`
	UpdatedAt                    *time.Time          `xml:"updated-at"`
	AuthorizationExpiresAt       *time.Time          `xml:"authorization-expires-at"`
	ThreeDSecureInfo             *ThreeDSecureInfo   `xml:"three-d-secure-info,omitempty"`
	Amount                       *Decimal            `xml:"amount"`
	SubscriptionDetails          *SubscriptionDetail `xml:"subscription"`
	CreditCard                   *CreditCard         `xml:"credit-card"`
	Customer                     *Customer           `xml:"customer"`
	BillingAddress               *Address            `xml:"billing"`
	ShippingAddress              *Address            `xml:"shipping"`
	TaxAmount                    *Decimal            `xml:"tax-amount"`
	ServiceFeeAmount             *Decimal            `xml:"service-fee-amount,attr"`
	DisbursementDetails          *DisbursementDetail `xml:"disbursement-details"`
	PayPalDetails                *PayPalDetail       `xml:"paypal"`
	VenmoAccountDetails          *VenmoAccountDetail `xml:"venmo-account"`
	AndroidPayDetails            *AndroidPayDetail   `xml:"android-pay-card"`
	ApplePayDetails              *ApplePayDetail     `xml:"apple-pay"`
	RiskData                     *RiskData           `xml:"risk-data"`
	Descriptor                   *Descriptor         `xml:"descriptor"`
	RefundedTransactionId        *string             `xml:"refunded-transaction-id"`
	RefundIds                    *[]string           `xml:"refund-ids>item"`
	Disputes                     []*Dispute          `xml:"disputes>dispute"`
}

type TxRequest struct {
	XMLName             string           `xml:"transaction"`
	CustomerID          string           `xml:"customer-id,omitempty"`
	Type                string           `xml:"type,omitempty"`
	OrderId             string           `xml:"order-id,omitempty"`
	PaymentMethodToken  string           `xml:"payment-method-token,omitempty"`
	PaymentMethodNonce  string           `xml:"payment-method-nonce,omitempty"`
	MerchantAccountId   string           `xml:"merchant-account-id,omitempty"`
	PlanId              string           `xml:"plan-id,omitempty"`
	DeviceData          string           `xml:"device-data,omitempty"`
	Channel             string           `xml:"channel,omitempty"`
	PurchaseOrderNumber string           `xml:"purchase-order-number,omitempty"`
	TaxExempt           bool             `xml:"tax-exempt,omitempty"`
	CustomFields        CustomFields     `xml:"custom-fields,omitempty"`
	TransactionSource   TxSource         `xml:"transaction-source,omitempty"`
	LineItems           LineItemRequests `xml:"line-items,omitempty"`
	Amount              *Decimal         `xml:"amount"`
	CreditCard          *CreditCard      `xml:"credit-card,omitempty"`
	Customer            *CustomerRequest `xml:"customer,omitempty"`
	BillingAddress      *Address         `xml:"billing,omitempty"`
	ShippingAddress     *Address         `xml:"shipping,omitempty"`
	TaxAmount           *Decimal         `xml:"tax-amount,omitempty"`
	Options             *TxOpts          `xml:"options,omitempty"`
	ServiceFeeAmount    *Decimal         `xml:"service-fee-amount,attr,omitempty"`
	RiskData            *RiskDataRequest `xml:"risk-data,omitempty"`
	Descriptor          *Descriptor      `xml:"descriptor,omitempty"`
}

type RefundRequest struct {
	XMLName string   `xml:"transaction"`
	Amount  *Decimal `xml:"amount"`
	OrderID string   `xml:"order-id,omitempty"`
}

func (t *Tx) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type typeWithNoFunctions Tx
	if err := d.DecodeElement((*typeWithNoFunctions)(t), &start); err != nil {
		return err
	}
	if t.SubscriptionDetails != nil &&
		t.SubscriptionDetails.BillingPeriodStartDate == "" &&
		t.SubscriptionDetails.BillingPeriodEndDate == "" {
		t.SubscriptionDetails = nil
	}
	return nil
}

type Transactions struct {
	Transaction []*Tx `xml:"transaction"`
}

type TxOpts struct {
	SubmitForSettlement              bool                       `xml:"submit-for-settlement,omitempty"`
	StoreInVault                     bool                       `xml:"store-in-vault,omitempty"`
	StoreInVaultOnSuccess            bool                       `xml:"store-in-vault-on-success,omitempty"`
	AddBillingAddressToPaymentMethod bool                       `xml:"add-billing-address-to-payment-method,omitempty"`
	StoreShippingAddressInVault      bool                       `xml:"store-shipping-address-in-vault,omitempty"`
	HoldInEscrow                     bool                       `xml:"hold-in-escrow,omitempty"`
	SkipAdvancedFraudChecking        bool                       `xml:"skip_advanced_fraud_checking,omitempty"`
	TransactionOptionsPaypalRequest  *TxPaypalOptsRequest       `xml:"paypal,omitempty"`
	ThreeDSecure                     *TxThreeDSecureOptsRequest `xml:"three-d-secure,omitempty"`
}

type TxPaypalOptsRequest struct {
	CustomField       string
	PayeeEmail        string
	Description       string
	SupplementaryData map[string]string
}

func (r TxPaypalOptsRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if r.CustomField != "" {
		if err := e.EncodeElement(r.CustomField, xml.StartElement{Name: xml.Name{Local: "custom-field"}}); err != nil {
			return err
		}
	}
	if r.PayeeEmail != "" {
		if err := e.EncodeElement(r.PayeeEmail, xml.StartElement{Name: xml.Name{Local: "payee-email"}}); err != nil {
			return err
		}
	}
	if r.Description != "" {
		if err := e.EncodeElement(r.Description, xml.StartElement{Name: xml.Name{Local: "description"}}); err != nil {
			return err
		}
	}
	if len(r.SupplementaryData) > 0 {
		start := xml.StartElement{Name: xml.Name{Local: "supplementary-data"}}
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		for k, v := range r.SupplementaryData {
			if err := e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
				return err
			}
		}
		if err := e.EncodeToken(start.End()); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	if err := e.Flush(); err != nil {
		return err
	}

	return nil
}

type TxThreeDSecureOptsRequest struct {
	Required bool `xml:"required"`
}

type TransactionSearchResult struct {
	TotalItems        int
	CurrentPageNumber int
	PageSize          int
	TotalIDs          []string
	Transactions      []*Tx
}

type RiskData struct {
	ID       string `xml:"id"`
	Decision string `xml:"decision"`
}

type RiskDataRequest struct {
	CustomerBrowser string `xml:"customer-browser"`
	CustomerIP      string `xml:"customer-ip"`
}

type SubscriptionDetail struct {
	BillingPeriodStartDate string `xml:"billing-period-start-date"`
	BillingPeriodEndDate   string `xml:"billing-period-end-date"`
}

func (c *APIClient) Pay(ctx context.Context, tx *TxRequest) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPost, transactionsPath, tx)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

type TxCloneRequest struct {
	XMLName string       `xml:"transaction-clone"`
	Amount  *Decimal     `xml:"amount"`
	Channel string       `xml:"channel"`
	Options *TxCloneOpts `xml:"options"`
}

type TxCloneOpts struct {
	SubmitForSettlement bool `xml:"submit-for-settlement"`
}

func (c *APIClient) Clone(ctx context.Context, id string, tx *TxCloneRequest) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPost, transactionsPath+"/"+id+"/clone", tx)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 201:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) SubmitForSettlement(ctx context.Context, id string, amounts ...*Decimal) (*Tx, error) {
	var tx *TxRequest
	if len(amounts) > 0 {
		tx = &TxRequest{
			Amount: amounts[0],
		}
	}
	response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+id+"/submit_for_settlement", tx)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}
func (c *APIClient) Void(ctx context.Context, id string) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+id+"/void", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) CancelRelease(ctx context.Context, id string) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+id+"/cancel_release", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) ReleaseFromEscrow(ctx context.Context, id string) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+id+"/release_from_escrow", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) HoldInEscrow(ctx context.Context, id string) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPut, transactionsPath+"/"+id+"/hold_in_escrow", nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) Refund(ctx context.Context, id string, amount ...*Decimal) (*Tx, error) {
	var tx *TxRequest
	if len(amount) > 0 {
		tx = &TxRequest{
			Amount: amount[0],
		}
	}
	response, err := c.do(ctx, http.MethodPost, transactionsPath+"/"+id+"/refund", tx)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200, 201:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) RefundWithRequest(ctx context.Context, id string, request *RefundRequest) (*Tx, error) {
	response, err := c.do(ctx, http.MethodPost, transactionsPath+"/"+id+"/refund", request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200, 201:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) FindTransaction(ctx context.Context, id string) (*Tx, error) {
	response, err := c.do(ctx, http.MethodGet, transactionsPath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Tx
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) SearchTxs(ctx context.Context, query *Search) (*SearchResult, error) {
	response, err := c.do(ctx, http.MethodPost, transactionsPath+"/"+advancedSearchIdsPath, query)
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

func (c *APIClient) SearchTx(ctx context.Context, query *Search, searchResult *SearchResult) (*TransactionSearchResult, error) {
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
	transactions, err := c.FetchTx(ctx, pageQuery)

	pageResult := &TransactionSearchResult{
		TotalItems:        len(searchResult.IDs),
		TotalIDs:          searchResult.IDs,
		CurrentPageNumber: searchResult.Page,
		PageSize:          searchResult.PageSize,
		Transactions:      transactions,
	}

	return pageResult, err
}

func (c *APIClient) FetchTx(ctx context.Context, query *Search) ([]*Tx, error) {
	response, err := c.do(ctx, http.MethodPost, transactionsPath+"/"+advancedSearchPath, query)
	if err != nil {
		return nil, err
	}
	var v struct {
		XMLName      string `xml:"credit-card-transactions"`
		Transactions []*Tx  `xml:"transaction"`
	}
	err = xml.Unmarshal(response.Body, &v)
	if err != nil {
		return nil, err
	}
	return v.Transactions, err
}

type testOperationPerformedInProductionError struct {
	error
}

func (e *testOperationPerformedInProductionError) Error() string {
	return fmt.Sprint("Operation not allowed in production environment")
}
