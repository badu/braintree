package braintree

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"
)

const disputesPath = "disputes"

type DisputeCategory string

const (
	DisputeDeviceId                                   DisputeCategory = "DEVICE_ID"
	DisputeDeviceName                                 DisputeCategory = "DEVICE_NAME"
	DisputePriorDigitalGoodsTransactionArn            DisputeCategory = "PRIOR_DIGITAL_GOODS_TRANSACTION_ARN"
	DisputePriorDigitalGoodsTransactionDateTime       DisputeCategory = "PRIOR_DIGITAL_GOODS_TRANSACTION_DATE_TIME"
	DisputeDownloadDateTime                           DisputeCategory = "DOWNLOAD_DATE_TIME"
	DisputeGeographicalLocation                       DisputeCategory = "GEOGRAPHICAL_LOCATION"
	DisputeLegitPaymentsForSameMerchandise            DisputeCategory = "LEGIT_PAYMENTS_FOR_SAME_MERCHANDISE"
	DisputeMerchantWebsiteOrAppAccess                 DisputeCategory = "MERCHANT_WEBSITE_OR_APP_ACCESS"
	DisputePriorNonDisputedTransactionArn             DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_ARN"
	DisputePriorNonDisputedTransactionDateTime        DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_DATE_TIME"
	DisputePriorNonDisputedTransactionEmailAddress    DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_EMAIL_ADDRESS"
	DisputePriorNonDisputedTransactionIpAddress       DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_IP_ADDRESS"
	DisputePriorNonDisputedTransactionPhoneNumber     DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_PHONE_NUMBER"
	DisputePriorNonDisputedTransactionPhysicalAddress DisputeCategory = "PRIOR_NON_DISPUTED_TRANSACTION_PHYSICAL_ADDRESS"
	DisputeProfileSetupOrAppAccess                    DisputeCategory = "PROFILE_SETUP_OR_APP_ACCESS"
	DisputeProofOfAuthorizedSigner                    DisputeCategory = "PROOF_OF_AUTHORIZED_SIGNER"
	DisputeProofOfDeliveryEmpAddress                  DisputeCategory = "PROOF_OF_DELIVERY_EMP_ADDRESS"
	DisputeProofOfDelivery                            DisputeCategory = "PROOF_OF_DELIVERY"
	DisputeProofOfPossessionOrUsage                   DisputeCategory = "PROOF_OF_POSSESSION_OR_USAGE"
	DisputePurchaserEmailAddress                      DisputeCategory = "PURCHASER_EMAIL_ADDRESS"
	DisputePurchaserIpAddress                         DisputeCategory = "PURCHASER_IP_ADDRESS"
	DisputePurchaserName                              DisputeCategory = "PURCHASER_NAME"
	DisputeRecurringTransactionArn                    DisputeCategory = "RECURRING_TRANSACTION_ARN"
	DisputeRecurringTransactionDateTime               DisputeCategory = "RECURRING_TRANSACTION_DATE_TIME"
	DisputeSignedDeliveryForm                         DisputeCategory = "SIGNED_DELIVERY_FORM"
	DisputeSignedOrderForm                            DisputeCategory = "SIGNED_ORDER_FORM"
	DisputeTicketProof                                DisputeCategory = "TICKET_PROOF"
)

type DisputeEvidence struct {
	XMLName           string          `xml:"evidence"`
	Comment           string          `xml:"comment"`
	CreatedAt         *time.Time      `xml:"created-at"`
	ID                string          `xml:"id"`
	SentToProcessorAt string          `xml:"sent-to-processor-at"`
	URL               string          `xml:"url"`
	Category          DisputeCategory `xml:"category"`
	SequenceNumber    string          `xml:"sequence-number"`
}

type DisputeTextEvidenceRequest struct {
	XMLName        xml.Name        `xml:"evidence"`
	Content        string          `xml:"comments"`
	Category       DisputeCategory `xml:"category,omitempty"`
	SequenceNumber string          `xml:"sequence-number,omitempty"`
}

type DisputeKind string

const (
	DisputeChargeback     DisputeKind = "chargeback"
	DisputePreArbitration DisputeKind = "pre_arbitration"
	DisputeRetrieval      DisputeKind = "retrieval"
)

type DisputeReason string

const (
	CancelledRecurringTransactionReason DisputeReason = "cancelled_recurring_transaction"
	CreditNotProcessedReason            DisputeReason = "credit_not_processed"
	DuplicateDisputeReason              DisputeReason = "duplicate"
	FraudDisputeReason                  DisputeReason = "fraud"
	GeneralReason                       DisputeReason = "general"
	InvalidAccountReason                DisputeReason = "invalid_account"
	NotRecognizedReason                 DisputeReason = "not_recognized"
	ProductNotReceivedReason            DisputeReason = "product_not_received"
	ProductUnsatisfactoryReason         DisputeReason = "product_unsatisfactory"
	TransactionAmountDiffersReason      DisputeReason = "transaction_amount_differs"
)

type DisputeStatus string

const (
	DisputeStatusAccepted DisputeStatus = "accepted"
	DisputeStatusDisputed DisputeStatus = "disputed"
	DisputeStatusExpired  DisputeStatus = "expired"
	DisputeStatusOpen     DisputeStatus = "open"
	DisputeStatusLost     DisputeStatus = "lost"
	DisputeStatusWon      DisputeStatus = "won"
)

type Dispute struct {
	XMLName           string                       `xml:"dispute"`
	CaseNumber        string                       `xml:"case-number"`
	CurrencyISOCode   string                       `xml:"currency-iso-code"`
	ID                string                       `xml:"id"`
	MerchantAccountID string                       `xml:"merchant-account-id"`
	OriginalDisputeID string                       `xml:"original-dispute-id"`
	ProcessorComments string                       `xml:"processor-comments"`
	ReturnCode        string                       `xml:"return-code"`
	ReceivedDate      string                       `xml:"received-date"`
	ReferenceNumber   string                       `xml:"reference-number"`
	ReplyByDate       string                       `xml:"reply-by-date"`
	Kind              DisputeKind                  `xml:"kind"`
	Reason            DisputeReason                `xml:"reason"`
	Status            DisputeStatus                `xml:"status"`
	AmountDisputed    *Decimal                     `xml:"amount-disputed"`
	AmountWon         *Decimal                     `xml:"amount-won"`
	UpdatedAt         *time.Time                   `xml:"updated-at"`
	CreatedAt         *time.Time                   `xml:"created-at"`
	Evidence          []*DisputeEvidence           `xml:"evidence>evidence"`
	StatusHistory     []*DisputeStatusHistoryEvent `xml:"status-history>status-history"`
	Transaction       *DisputeTransaction          `xml:"transaction"`
}

type DisputeStatusHistoryEvent struct {
	XMLName          string     `xml:"status-history"`
	DisbursementDate string     `xml:"disbursement-date"`
	EffectiveDate    string     `xml:"effective-date"`
	Status           string     `xml:"status"`
	Timestamp        *time.Time `xml:"timestamp"`
}

type DisputeTransaction struct {
	XMLName                  string     `xml:"transaction"`
	ID                       string     `xml:"id"`
	OrderID                  string     `xml:"order-id"`
	PaymentInstrumentSubtype string     `xml:"payment-instrument-subtype"`
	PurchaseOrderNumber      string     `xml:"purchase-order-number"`
	Amount                   *Decimal   `xml:"amount"`
	CreatedAt                *time.Time `xml:"created-at"`
}

func (c *APIClient) FindDispute(ctx context.Context, disputeID string) (*Dispute, error) {
	response, err := c.call(ctx, http.MethodGet, disputesPath+"/"+disputeID, nil, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result Dispute
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) AddTextEvidence(ctx context.Context, disputeID string, evidence *DisputeTextEvidenceRequest) (*DisputeEvidence, error) {
	response, err := c.call(ctx, http.MethodPost, disputesPath+"/"+disputeID+"/evidence", evidence, apiVersion4)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		var result DisputeEvidence
		if err := xml.Unmarshal(response.Body, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, &invalidResponseError{response}
}

func (c *APIClient) RemoveEvidence(ctx context.Context, disputeID string, id string) error {
	resp, err := c.call(ctx, http.MethodDelete, disputesPath+"/"+disputeID+"/evidence/"+id, nil, apiVersion4)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}

func (c *APIClient) Accept(ctx context.Context, disputeID string) error {
	resp, err := c.call(ctx, http.MethodPut, disputesPath+"/"+disputeID+"/accept", nil, apiVersion4)
	if err != nil {
		return nil
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}

func (c *APIClient) Finalize(ctx context.Context, disputeID string) error {
	resp, err := c.call(ctx, http.MethodPut, disputesPath+"/"+disputeID+"/finalize", nil, apiVersion4)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	}
	return &invalidResponseError{resp}
}
