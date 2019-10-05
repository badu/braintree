// +build unit

package tests

import (
	"encoding/xml"
	"testing"
	"time"

	. "github.com/badu/braintree"
)

func TestClientTokenMarshalXML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *TokenRequest
		wantXML string
	}{
		{
			name:    "request nil",
			req:     nil,
			wantXML: ``,
		},
		{
			name: "request empty",
			req:  &TokenRequest{},
			wantXML: `<client-token>
  <version>0</version>
</client-token>`,
		},
		{
			name: "request with provided version",
			req:  &TokenRequest{Version: 2},
			wantXML: `<client-token>
  <version>2</version>
</client-token>`,
		},
		{
			name: "request with customer and merchant account",
			req: &TokenRequest{
				CustomerID:        "1234",
				MerchantAccountID: "5678",
			},
			wantXML: `<client-token>
  <customer-id>1234</customer-id>
  <merchant-account-id>5678</merchant-account-id>
  <version>0</version>
</client-token>`,
		},
		{
			name: "request with non-pointer options false",
			req: &TokenRequest{
				Options: &Options{
					FailOnDuplicatePaymentMethod: false,
					MakeDefault:                  false,
				},
			},
			wantXML: `<client-token>
  <options></options>
  <version>0</version>
</client-token>`,
		},
		{
			name: "request with non-pointer options true",
			req: &TokenRequest{
				Options: &Options{
					FailOnDuplicatePaymentMethod: true,
					MakeDefault:                  true,
				},
			},
			wantXML: `<client-token>
  <options>
    <fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
    <make-default>true</make-default>
  </options>
  <version>0</version>
</client-token>`,
		},
		{
			name: "request with verify card true",
			req: &TokenRequest{
				Options: &Options{
					VerifyCard: BoolPtr(true),
				},
			},
			wantXML: `<client-token>
  <options>
    <verify-card>true</verify-card>
  </options>
  <version>0</version>
</client-token>`,
		},
		{
			name: "request with verify card false",
			req: &TokenRequest{
				Options: &Options{
					VerifyCard: BoolPtr(false),
				},
			},
			wantXML: `<client-token>
  <options>
    <verify-card>false</verify-card>
  </options>
  <version>0</version>
</client-token>`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := xml.MarshalIndent(test.req, "", "  ")
			xml := string(output)
			if err != nil {
				t.Fatalf("got error = %v", err)
			}
			if xml != test.wantXML {
				t.Errorf("got xml:\n%s\nwant xml:\n%s", xml, test.wantXML)
			}
		})
	}
}

func TestCreditCardOptionsMarshalXML(t *testing.T) {
	bTrue := true
	bFalse := false
	tests := []struct {
		name    string
		cco     *CreditCardOptions
		wantXML string
		wantErr bool
	}{
		{
			name:    "nil pointer",
			cco:     nil,
			wantXML: ``,
			wantErr: false,
		},
		{
			name: "VerifyCard nil",
			cco: &CreditCardOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
			},
			wantXML: `<CreditCardOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
</CreditCardOptions>`,
			wantErr: false,
		},
		{
			name: "VerifyCard true",
			cco: &CreditCardOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bTrue,
			},
			wantXML: `<CreditCardOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<verify-card>true</verify-card>
</CreditCardOptions>`,
			wantErr: false,
		},
		{
			name: "VerifyCard false",
			cco: &CreditCardOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bFalse,
			},
			wantXML: `<CreditCardOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<verify-card>false</verify-card>
</CreditCardOptions>`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, err := xml.MarshalIndent(tt.cco, "", "\t")
			xml := string(output)
			if err != nil {
				t.Fatalf("got error = %v", err)
			}
			if xml != tt.wantXML {
				t.Errorf("got xml:\n%s\nwant xml:\n%s", xml, tt.wantXML)
			}
		})
	}
}

func TestPaymentMethodRequestOptionsMarshalXML(t *testing.T) {
	bTrue := true
	bFalse := false
	tests := []struct {
		name    string
		pmo     *PaymentMethodRequestOptions
		wantXML string
		wantErr bool
	}{
		{
			name:    "nil pointer",
			pmo:     nil,
			wantXML: ``,
			wantErr: false,
		},
		{
			name: "VerifyCard nil",
			pmo: &PaymentMethodRequestOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
			},
			wantXML: `<PaymentMethodRequestOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
</PaymentMethodRequestOptions>`,
			wantErr: false,
		},
		{
			name: "VerifyCard true",
			pmo: &PaymentMethodRequestOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bTrue,
			},
			wantXML: `<PaymentMethodRequestOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<verify-card>true</verify-card>
</PaymentMethodRequestOptions>`,
			wantErr: false,
		},
		{
			name: "VerifyCard false",
			pmo: &PaymentMethodRequestOptions{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bFalse,
			},
			wantXML: `<PaymentMethodRequestOptions>
	<make-default>true</make-default>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<verify-card>false</verify-card>
</PaymentMethodRequestOptions>`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, err := xml.MarshalIndent(tt.pmo, "", "\t")
			xml := string(output)
			if err != nil {
				t.Fatalf("got error = %v", err)
			}
			if xml != tt.wantXML {
				t.Errorf("got xml:\n%s\nwant xml:\n%s", xml, tt.wantXML)
			}
		})
	}
}

func TestClientTokenRequestOptionsMarshalXML(t *testing.T) {
	bTrue := true
	bFalse := false
	tests := []struct {
		name    string
		ctro    *Options
		wantXML string
		wantErr bool
	}{
		{
			name:    "nil pointer",
			ctro:    nil,
			wantXML: ``,
			wantErr: false,
		},
		{
			name: "VerifyCard nil",
			ctro: &Options{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
			},
			wantXML: `<Options>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<make-default>true</make-default>
</Options>`,
			wantErr: false,
		},
		{
			name: "VerifyCard true",
			ctro: &Options{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bTrue,
			},
			wantXML: `<Options>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<make-default>true</make-default>
	<verify-card>true</verify-card>
</Options>`,
			wantErr: false,
		},
		{
			name: "VerifyCard false",
			ctro: &Options{
				FailOnDuplicatePaymentMethod: true,
				MakeDefault:                  true,
				VerifyCard:                   &bFalse,
			},
			wantXML: `<Options>
	<fail-on-duplicate-payment-method>true</fail-on-duplicate-payment-method>
	<make-default>true</make-default>
	<verify-card>false</verify-card>
</Options>`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, err := xml.MarshalIndent(tt.ctro, "", "\t")
			xml := string(output)
			if err != nil {
				t.Fatalf("got error = %v", err)
			}
			if xml != tt.wantXML {
				t.Errorf("got xml:\n%s\nwant xml:\n%s", xml, tt.wantXML)
			}
		})
	}
}

func TestDecimalMarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  *Decimal
		out []byte
	}{
		{NewDecimal(250, -2), []byte("25000")},
		{NewDecimal(2, 0), []byte("2")},
		{NewDecimal(23, 0), []byte("23")},
		{NewDecimal(234, 0), []byte("234")},
		{NewDecimal(0, 1), []byte("0.0")},
		{NewDecimal(1, 1), []byte("0.1")},
		{NewDecimal(12, 1), []byte("1.2")},
		{NewDecimal(0, 2), []byte("0.00")},
		{NewDecimal(5, 2), []byte("0.05")},
		{NewDecimal(55, 2), []byte("0.55")},
		{NewDecimal(250, 2), []byte("2.50")},
		{NewDecimal(4586, 2), []byte("45.86")},
		{NewDecimal(-5504, 2), []byte("-55.04")},
		{NewDecimal(0, 3), []byte("0.000")},
		{NewDecimal(5, 3), []byte("0.005")},
		{NewDecimal(55, 3), []byte("0.055")},
		{NewDecimal(250, 3), []byte("0.250")},
		{NewDecimal(4586, 3), []byte("4.586")},
		{NewDecimal(45867, 3), []byte("45.867")},
		{NewDecimal(-55043, 3), []byte("-55.043")},
	}

	for _, tt := range tests {
		b, err := tt.in.MarshalText()
		if err != nil {
			t.Errorf("expected %+v.MarshalText() => to not error, but it did with %s", tt.in, err)
		}
		if string(tt.out) != string(b) {
			t.Errorf("%+v.MarshalText() => %s, want %s", tt.in, b, tt.out)
		}
	}
}

func TestSearchXMLEncode(t *testing.T) {
	t.Parallel()

	s := new(Search)

	f := s.AddTextField("customer-first-name")
	f.Is = "A"
	f.IsNot = "B"
	f.StartsWith = "C"
	f.EndsWith = "D"
	f.Contains = "E"

	f2 := s.AddRangeField("amount")
	f2.Is = 15.01
	f2.Min = 10.01
	f2.Max = 20.01

	startDate := time.Date(2016, time.September, 11, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2016, time.September, 11, 23, 59, 59, 0, time.UTC)
	f3 := s.AddTimeField("settled-at")
	f3.Min = startDate
	f3.Max = endDate

	f4 := s.AddTimeField("created-at")
	f4.Min = startDate

	f5 := s.AddTimeField("authorization-expired-at")
	f5.Min = startDate

	f6 := s.AddTimeField("authorized-at")
	f6.Min = startDate

	f7 := s.AddTimeField("failed-at")
	f7.Min = startDate

	f8 := s.AddTimeField("gateway-rejected-at")
	f8.Min = startDate

	f9 := s.AddTimeField("processor-declined-at")
	f9.Min = startDate

	f10 := s.AddTimeField("submitted-for-settlement-at")
	f10.Min = startDate

	f11 := s.AddTimeField("voided-at")
	f11.Min = startDate

	f12 := s.AddTimeField("disbursement-date")
	f12.Min = startDate

	f13 := s.AddTimeField("dispute-date")
	f13.Min = startDate

	f14 := s.AddMultiField("status")
	f14.Items = []string{
		string(StatusAuthorized),
		string(StatusSubmittedForSettlement),
		string(StatusSettled),
	}

	b, err := xml.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	xmls := string(b)

	expect := `<search>
  <customer-first-name>
    <is>A</is>
    <is-not>B</is-not>
    <starts-with>C</starts-with>
    <ends-with>D</ends-with>
    <contains>E</contains>
  </customer-first-name>
  <amount>
    <is>15.01</is>
    <min>10.01</min>
    <max>20.01</max>
  </amount>
  <settled-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
    <max type="datetime">2016-09-11T23:59:59Z</max>
  </settled-at>
  <created-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </created-at>
  <authorization-expired-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </authorization-expired-at>
  <authorized-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </authorized-at>
  <failed-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </failed-at>
  <gateway-rejected-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </gateway-rejected-at>
  <processor-declined-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </processor-declined-at>
  <submitted-for-settlement-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </submitted-for-settlement-at>
  <voided-at>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </voided-at>
  <disbursement-date>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </disbursement-date>
  <dispute-date>
    <min type="datetime">2016-09-11T00:00:00Z</min>
  </dispute-date>
  <status type="array">
    <item>authorized</item>
    <item>submitted_for_settlement</item>
    <item>settled</item>
  </status>
</search>`

	if xmls != expect {
		t.Fatalf("got %#v, want %#v", xmls, expect)
	}
}
