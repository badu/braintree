package braintree

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

const (
	CheckWH                             = "check"
	DisbursementWH                      = "disbursement"
	DisbursementExceptionWH             = "disbursement_exception"
	SubscriptionCanceledWH              = "subscription_canceled"
	SubscriptionChargedSuccessfullyWH   = "subscription_charged_successfully"
	SubscriptionChargedUnsuccessfullyWH = "subscription_charged_unsuccessfully"
	SubscriptionExpiredWH               = "subscription_expired"
	SubscriptionTrialEndedWH            = "subscription_trial_ended"
	SubscriptionWentActiveWH            = "subscription_went_active"
	SubscriptionWentPastDueWH           = "subscription_went_past_due"
	SubMerchantAccountApprovedWH        = "sub_merchant_account_approved"
	SubMerchantAccountDeclinedWH        = "sub_merchant_account_declined"
	PartnerMerchantConnectedWH          = "partner_merchant_connected"
	PartnerMerchantDisconnectedWH       = "partner_merchant_disconnected"
	PartnerMerchantDeclinedWH           = "partner_merchant_declined"
	TransactionSettledWH                = "transaction_settled"
	TransactionSettlementDeclinedWH     = "transaction_settlement_declined"
	TransactionDisbursedWH              = "transaction_disbursed"
	DisputeOpenedWH                     = "dispute_opened"
	DisputeLostWH                       = "dispute_lost"
	DisputeWonWH                        = "dispute_won"
	AccountUpdaterDailyReportWH         = "account_updater_daily_report"
)

type Notification struct {
	XMLName   xml.Name  `xml:"notification"`
	Kind      string    `xml:"kind"`
	Timestamp time.Time `xml:"timestamp"`
	Subject   *Subject  `xml:"subject"`
}

func (n *Notification) MerchantAccount() *MerchantAccount {
	if n.Subject.APIErrorResponse != nil && n.Subject.APIErrorResponse.MerchantAccount != nil {
		return n.Subject.APIErrorResponse.MerchantAccount
	} else if n.Subject.MerchantAccount != nil {
		return n.Subject.MerchantAccount
	}
	return nil
}

func (n *Notification) Disbursement() *Disbursement {
	if n.Subject.Disbursement != nil {
		return n.Subject.Disbursement
	}
	return nil
}

func (n *Notification) Dispute() *Dispute {
	if n.Subject.Dispute != nil {
		return n.Subject.Dispute
	}
	return nil
}

func (n *Notification) AccountUpdaterDailyReport() *DailyReport {
	if n.Subject.AccountUpdaterDailyReport != nil {
		return n.Subject.AccountUpdaterDailyReport
	}
	return nil
}

type DailyReport struct {
	XMLName    xml.Name `xml:"account-updater-daily-report"`
	ReportDate string   `xml:"report-date"`
	ReportURL  string   `xml:"report-url"`
}

type Disbursement struct {
	XMLName          xml.Name         `xml:"disbursement"`
	Id               string           `xml:"id"`
	ExceptionMessage string           `xml:"exception-message"`
	CurrencyIsoCode  string           `xml:"currency-iso-code"`
	Status           string           `xml:"status"`
	FollowUpAction   string           `xml:"follow-up-action"`
	Success          bool             `xml:"success"`
	Retry            bool             `xml:"retry"`
	IsSubmerchant    bool             `xml:"sub-merchant-account"`
	TransactionIds   []string         `xml:"transaction-ids>item"`
	DisbursementDate *Date            `xml:"disbursement-date"`
	Amount           *Decimal         `xml:"amount"`
	MerchantAccount  *MerchantAccount `xml:"merchant-account"`
}

const (
	// Exception messages
	BankRejected         = "bank_rejected"
	InsufficientFunds    = "insuffient_funds"
	AccountNotAuthorized = "account_not_authorized"

	// Followup actions
	ContactUs                = "contact_us"
	UpdateFundingInformation = "update_funding_information"
	None                     = "none"
)

func (d *Disbursement) Transactions(ctx context.Context, c *APIClient) (*TransactionSearchResult, error) {
	query := new(Search)
	f := query.AddMultiField("ids")
	f.Items = d.TransactionIds
	searchResult, err := c.SearchTxs(ctx, query)
	if err != nil {
		return nil, err
	}

	pageSize := searchResult.PageSize
	ids := searchResult.IDs

	endOffset := pageSize
	if endOffset > len(ids) {
		endOffset = len(ids)
	}

	firstPageQuery := query.ShallowCopy()
	firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
	firstPageTransactions, err := c.FetchTx(ctx, firstPageQuery)

	result := &TransactionSearchResult{
		TotalItems:        len(ids),
		TotalIDs:          ids,
		CurrentPageNumber: 1,
		PageSize:          pageSize,
		Transactions:      firstPageTransactions,
	}

	return result, err
}

type Subject struct {
	XMLName                   xml.Name         `xml:"subject"`
	APIErrorResponse          *APIError        `xml:"api-error-response,omitempty"`
	Disbursement              *Disbursement    `xml:"disbursement,omitempty"`
	Subscription              *Subscription    `xml:",omitempty"`
	MerchantAccount           *MerchantAccount `xml:"merchant-account,omitempty"`
	Transaction               *Tx              `xml:",omitempty"`
	Dispute                   *Dispute         `xml:"dispute,omitempty"`
	AccountUpdaterDailyReport *DailyReport     `xml:"account-updater-daily-report,omitempty"`
}

type SignatureError struct {
	Message string
}

func (i SignatureError) Error() string {
	if i.Message == "" {
		return "Invalid Signature"
	}
	return i.Message
}

func NewKey(publicKey, privateKey string) Key {
	return Key{PublicKey: publicKey, PrivateKey: privateKey}
}

type Key struct {
	PublicKey  string
	PrivateKey string
}

func (k Key) VerifySignature(signature, payload string) (bool, error) {
	signature, err := k.ParseSignature(signature)
	if err != nil {
		return false, err
	}
	expectedSignature, err := k.HMAC(payload)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expectedSignature), []byte(signature)), nil
}

func (k Key) getMatchingSignature(signaturePairs string) (sig string, ok bool) {
	pairs := strings.Split(signaturePairs, "&")
	for _, pair := range pairs {
		split := strings.Split(pair, "|")
		if len(split) == 2 && split[0] == k.PublicKey {
			return split[1], true
		}
	}
	return "", false
}

func (k Key) ParseSignature(signatureKeyPairs string) (string, error) {
	if !strings.Contains(signatureKeyPairs, "|") {
		return "", SignatureError{"Signature-key pair does not contain |"}
	}
	signature, ok := k.getMatchingSignature(signatureKeyPairs)
	if !ok {
		return "", SignatureError{"Signature-key pair contains the wrong public key!"}
	}
	return signature, nil
}

func (k Key) HMAC(payload string) (string, error) {
	s := sha1.New()
	_, err := io.WriteString(s, k.PrivateKey)
	if err != nil {
		return "", errors.New("could not write private key to SHA1")
	}
	mac := hmac.New(sha1.New, s.Sum(nil))
	_, err = mac.Write([]byte(payload))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", mac.Sum(nil)), nil
}

func (c *APIClient) ParseRequest(r *http.Request) (*Notification, error) {
	signature := r.PostFormValue("bt_signature")
	payload := r.PostFormValue("bt_payload")
	return c.Parse(signature, payload)
}

func (c *APIClient) Parse(signature, payload string) (*Notification, error) {
	key := NewKey(c.Key.PublicKey, c.Key.PrivateKey)
	if verified, err := key.VerifySignature(signature, payload); err != nil {
		return nil, err
	} else if !verified {
		return nil, SignatureError{}
	}

	xmlNotification, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	strippedBuf, err := StripNilElements(xmlNotification)
	if err != nil {
		return nil, err
	}

	var result Notification
	err = xml.Unmarshal(strippedBuf, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) Verify(challenge string) (string, error) {
	key := NewKey(c.Key.PublicKey, c.Key.PrivateKey)
	digest, err := key.HMAC(challenge)
	if err != nil {
		return ``, err
	}
	return key.PublicKey + `|` + digest, nil
}

const payloadTemplate = `<notification>
	<timestamp type="datetime">%s</timestamp>
	<kind>%s</kind>
	<subject>%s</subject>
</notification>
`

func (c *APIClient) SandboxRequest(kind, id string) (*http.Request, error) {
	payload := c.SamplePayload(kind, id)
	signature, err := c.SignPayload(payload)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("bt_signature", signature)
	form.Add("bt_payload", payload)

	body := form.Encode()
	return &http.Request{
		Method:        http.MethodPost,
		Header:        http.Header{HdrContentType: {"application/x-www-form-urlencoded"}},
		ContentLength: int64(len(body)),
		Body:          ioutil.NopCloser(strings.NewReader(body)),
	}, nil
}

func (c *APIClient) SamplePayload(kind, id string) string {
	datetime := time.Now().UTC().Format(time.RFC3339)
	payload := fmt.Sprintf(
		payloadTemplate,
		datetime,
		kind,
		c.subjectXML(kind, id),
	)
	encodedpayload := base64.StdEncoding.EncodeToString([]byte(payload))
	return strings.Replace(encodedpayload, "\r", "", -1)
}

func (c *APIClient) SignPayload(src string) (string, error) {
	key := NewKey(c.Key.PublicKey, c.Key.PrivateKey)
	payload, err := key.HMAC(src)
	if err != nil {
		return "", err
	}
	return key.PublicKey + "|" + payload, nil
}

func (c *APIClient) subjectXML(kind, id string) string {
	type subjectXMLData struct {
		ID string
	}

	xmlTmpl := c.subjectXMLTemplate(kind)
	var result bytes.Buffer
	tmpl, err := template.New("").Parse(xmlTmpl)
	err = tmpl.Execute(&result, subjectXMLData{ID: id})
	if err != nil {
		panic(fmt.Errorf("creating xml template: " + err.Error()))
	}
	return result.String()
}

func (c *APIClient) subjectXMLTemplate(kind string) string {
	switch kind {
	case CheckWH:
		return c.checkXML()
	case SubMerchantAccountApprovedWH:
		return c.merchantAccountXMLApproved()
	case SubMerchantAccountDeclinedWH:
		return c.merchantAccountXMLDeclined()
	case TransactionDisbursedWH:
		return c.transactionDisbursedXML()
	case TransactionSettledWH:
		return c.transactionSettledXML()
	case TransactionSettlementDeclinedWH:
		return c.transactionSettlementDeclinedXML()
	case DisbursementWH:
		return c.disbursementXML()
	case DisputeOpenedWH:
		return c.disputeOpenedXML()
	case DisputeLostWH:
		return c.disputeLostXML()
	case DisputeWonWH:
		return c.disputeWonXML()
	case DisbursementExceptionWH:
		return c.disbursementExceptionXML()
	case PartnerMerchantConnectedWH:
		return c.partnerMerchantConnectedXML()
	case PartnerMerchantDisconnectedWH:
		return c.partnerMerchantDisconnectedXML()
	case PartnerMerchantDeclinedWH:
		return c.partnerMerchantDeclinedXML()
	case SubscriptionChargedSuccessfullyWH:
		return c.subscriptionChargedSuccessfullyXML()
	case AccountUpdaterDailyReportWH:
		return c.accountUpdaterDailyReportXML()
	default:
		return c.subscriptionXML()
	}
}

func (c *APIClient) checkXML() string {
	return `<check type="boolean">true</check>`
}

func (c *APIClient) merchantAccountXMLApproved() string {
	return `
		<merchant_account>
			<id>{{ $.ID }}</id>
			<master_merchant_account>
				<id>master_ma_for_{{ $.ID }}</id>
				<status>active</status>
			</master_merchant_account>
			<status>active</status>
		</merchant_account>
		`
}

func (c *APIClient) merchantAccountXMLDeclined() string {
	return `
		<api-error-response>
			<message>Credit score is too low</message>
			<errors type="array"/>
				<merchant-account>
					<errors type="array">
						<error>
							<code>82621</code>
							<message>Credit score is too low</message>
							<attribute type="symbol">base</attribute>
						</error>
					</errors>
				</merchant-account>
			</errors>
			<merchant-account>
				<id>{{ $.ID }}</id>
				<status>suspended</status>
				<master-merchant-account>
					<id>master_ma_for_{{ $.ID }}</id>
					<status>suspended</status>
				</master-merchant-account>
			</merchant-account>
		</api-error-response>
		`
}

func (c *APIClient) subscriptionXML() string {
	return `
		<subscription>
			<id>{{ $.ID }}</id>
			<transactions type="array">
			</transactions>
			<add_ons type="array">
			</add_ons>
			<discounts type="array">
			</discounts>
		</subscription>
		`
}

func (c *APIClient) subscriptionChargedSuccessfullyXML() string {
	return `
		<subscription>
			<id>{{ $.ID }}</id>
			<transactions type="array">
				<transaction>
					<id>{{ $.ID }}</id>
					<status>submitted_for_settlement</status>
					<amount>49.99</amount>
				</transaction>
			</transactions>
			<add_ons type="array">
			</add_ons>
			<discounts type="array">
			</discounts>
		</subscription>
		`
}

func (c *APIClient) transactionDisbursedXML() string {
	return `
		<transaction>
			<id>{{ $.ID }}</id>
			<amount>100</amount>
			<disbursement-details>
				<disbursement-date type="date">2013-07-09</disbursement-date>
			</disbursement-details>
		</transaction>
		`
}

func (c *APIClient) transactionSettledXML() string {
	return `
		<transaction>
			<id>{{ $.ID}}</id>
			<status>settled</status>
			<type>sale</type>
			<currency-iso-code>USD</currency-iso-code>
			<amount>100.00</amount>
			<merchant-account-id>ogaotkivejpfayqfeaimuktty</merchant-account-id>
			<payment-instrument-type>us_bank_account</payment-instrument-type>
			<us-bank-account>
				<routing-number>123456789</routing-number>
				<last-4>1234</last-4>
				<account-type>checking</account-type>
				<account-holder-name>Account Holder</account-holder-name>
			</us-bank-account>
		</transaction>
		`
}

func (c *APIClient) transactionSettlementDeclinedXML() string {
	return `
		<transaction>
			<id>{{ $.ID }}</id>
			<status>settlement_declined</status>
			<type>sale</type>
			<currency-iso-code>USD</currency-iso-code>
			<amount>100.00</amount>
			<merchant-account-id>ogaotkivejpfayqfeaimuktty</merchant-account-id>
			<payment-instrument-type>us_bank_account</payment-instrument-type>
			<us-bank-account>
				<routing-number>123456789</routing-number>
				<last-4>1234</last-4>
				<account-type>checking</account-type>
				<account-holder-name>Account Holder</account-holder-name>
			</us-bank-account>
		</transaction>
		`
}

func (c *APIClient) disputeOpenedXML() string {
	return `
		<dispute>
			<amount>250.00</amount>
			<currency-iso-code>USD</currency-iso-code>
			<received-date type="date">2020-05-01</received-date>
			<reply-by-date type="date">2020-06-01</reply-by-date>
			<kind>chargeback</kind>
			<status>open</status>
			<reason>fraud</reason>
			<id>{{ $.ID }}</id>
			<transaction>
				<id>{{ $.ID }}</id>
				<amount>250.00</amount>
			</transaction>
			<date-opened type="date">2020-06-01</date-opened>
		</dispute>
		`
}

func (c *APIClient) disputeLostXML() string {
	return `
		<dispute>
			<amount>250.00</amount>
			<currency-iso-code>USD</currency-iso-code>
			<received-date type="date">2020-05-01</received-date>
			<reply-by-date type="date">2020-06-01</reply-by-date>
			<kind>chargeback</kind>
			<status>lost</status>
			<reason>fraud</reason>
			<id>{{ $.ID }}</id>
			<transaction>
				<id>{{ $.ID }}</id>
				<amount>250.00</amount>
			</transaction>
			<date-opened type="date">2020-06-01</date-opened>
		</dispute>
		`
}

func (c *APIClient) disputeWonXML() string {
	return `
		<dispute>
			<amount>250.00</amount>
			<currency-iso-code>USD</currency-iso-code>
			<received-date type="date">2020-05-01</received-date>
			<reply-by-date type="date">2020-06-01</reply-by-date>
			<kind>chargeback</kind>
			<status>won</status>
			<reason>fraud</reason>
			<id>{{ $.ID }}</id>
			<transaction>
				<id>{{ $.ID }}</id>
				<amount>250.00</amount>
			</transaction>
			<date-opened type="date">2020-06-01</date-opened>
			<date-won type="date">2014-03-22</date-won>
		</dispute>
		`
}

func (c *APIClient) disbursementXML() string {
	return `
		<disbursement>
			<id>{{ $.ID }}</id>
			<transaction-ids type="array">
				<item>afv56j</item>
				<item>kj8hjk</item>
			</transaction-ids>
			<success type="boolean">true</success>
			<retry type="boolean">false</retry>
			<merchant-account>
				<id>merchant_account_token</id>
				<currency-iso-code>USD</currency-iso-code>
				<sub-merchant-account type="boolean">false</sub-merchant-account>
				<status>active</status>
			</merchant-account>
			<amount>100.00</amount>
			<disbursement-date type="date">2014-02-10</disbursement-date>
			<exception-message nil="true"/>
			<follow-up-action nil="true"/>
		</disbursement>
		`
}

func (c *APIClient) disbursementExceptionXML() string {
	return `
		<disbursement>
			<id>{{ $.ID }}</id>
			<transaction-ids type="array">
				<item>afv56j</item>
				<item>kj8hjk</item>
			</transaction-ids>
			<success type="boolean">false</success>
			<retry type="boolean">false</retry>
			<merchant-account>
				<id>merchant_account_token</id>
				<currency-iso-code>USD</currency-iso-code>
				<sub-merchant-account type="boolean">false</sub-merchant-account>
				<status>active</status>
			</merchant-account>
			<amount>100.00</amount>
			<disbursement-date type="date">2014-02-10</disbursement-date>
			<exception-message>bank_rejected</exception-message>
			<follow-up-action>update_funding_information</follow-up-action>
		</disbursement>
	`
}

func (c *APIClient) partnerMerchantConnectedXML() string {
	return `
	<partner-merchant>
		<merchant-public-id>public_id</merchant-public-id>
		<public-key>public_key</public-key>
		<private-key>private_key</private-key>
		<partner-merchant-id>goguSRL</partner-merchant-id>
		<client-side-encryption-key>cse_key</client-side-encryption-key>
	</partner-merchant>
	`
}

func (c *APIClient) partnerMerchantDisconnectedXML() string {
	return `
	<partner-merchant>
		<partner-merchant-id>goguSRL</partner-merchant-id>
	</partner-merchant>
	`
}

func (c *APIClient) partnerMerchantDeclinedXML() string {
	return `
	<partner-merchant>
		<partner-merchant-id>goguSRL</partner-merchant-id>
	</partner-merchant>
	`
}

func (c *APIClient) accountUpdaterDailyReportXML() string {
	return `
	<account-updater-daily-report>
		<report-date type="date">2020-01-01</report-date>
		<report-url>link-to-csv-report</report-url>
	</account-updater-daily-report>
	`
}
