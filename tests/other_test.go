// +build unit

package tests

import (
	"encoding/xml"
	"testing"

	. "github.com/badu/braintree"
)

func TestDecimalCmp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		x, y *Decimal
		out  int
	}{
		{NewDecimal(250, -2), NewDecimal(250, -2), 0},
		{NewDecimal(2, 0), NewDecimal(250, -2), -1},
		{NewDecimal(500, 2), NewDecimal(50, 1), 0},
		{NewDecimal(2500, -2), NewDecimal(250, -2), 1},
		{NewDecimal(100, 2), NewDecimal(1, 0), 0},
	}

	for i, tt := range tests {
		if out := tt.x.Cmp(tt.y); out != tt.out {
			t.Errorf("%d: %+v.Cmp(%+v) => %d, want %d", i, tt.x, tt.y, out, tt.out)
		}
	}
}

var errorXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<api-error-response>
  <errors>
    <errors type="array"/>
    <transaction>
      <errors type="array">
        <error>
          <code>91560</code>
          <attribute type="symbol">base</attribute>
          <message>Tx could not be held in escrow.</message>
        </error>
        <error>
          <code>81502</code>
          <attribute type="symbol">amount</attribute>
          <message>Amount is required.</message>
        </error>
        <error>
          <code>91526</code>
          <attribute type="symbol">custom_fields</attribute>
          <message>Custom field is invalid: store_me.</message>
        </error>
        <error>
          <code>91513</code>
          <attribute type="symbol">merchant_account_id</attribute>
          <message>Merchant account ID is invalid.</message>
        </error>
        <error>
          <code>915157</code>
          <attribute type="symbol">line_items</attribute>
          <message>Too many line items.</message>
        </error>
      </errors>
      <credit-card>
        <errors type="array">
          <error>
            <code>91708</code>
            <attribute type="symbol">base</attribute>
            <message>Cannot provide expiration_date if you are also providing expiration_month and expiration_year.</message>
          </error>
          <error>
            <code>81714</code>
            <attribute type="symbol">number</attribute>
            <message>Credit card number is required.</message>
          </error>
          <error>
            <code>81725</code>
            <attribute type="symbol">base</attribute>
            <message>Credit card must include either number or venmo_sdk_payment_method_code.</message>
          </error>
          <error>
            <code>81703</code>
            <attribute type="symbol">number</attribute>
            <message>Credit card type is not accepted by this merchant account.</message>
          </error>
        </errors>
      </credit-card>
      <customer>
        <errors type="array">
          <error>
            <code>81606</code>
            <attribute type="symbol">email</attribute>
            <message>Email is an invalid format.</message>
          </error>
        </errors>
      </customer>
      <line-items>
        <index-1>
          <errors type="array">
            <error>
              <code>95801</code>
              <attribute type="symbol">commodity_code</attribute>
              <message>Commodity code is too long.</message>
            </error>
          </errors>
        </index-1>
        <index-3>
          <errors type="array">
            <error>
              <code>95803</code>
              <attribute type="symbol">description</attribute>
              <message>Description is too long.</message>
            </error>
            <error>
              <code>95809</code>
              <attribute type="symbol">product_code</attribute>
              <message>Product code is too long.</message>
            </error>
          </errors>
        </index-3>
      </line-items>
    </transaction>
  </errors>
  <message>Everything is broken!</message>
</api-error-response>`)

func TestErrorsUnmarshalEverything(t *testing.T) {
	t.Parallel()

	apiErrors := &APIError{}
	err := xml.Unmarshal(errorXML, apiErrors)
	if err != nil {
		t.Fatal("Error unmarshalling: " + err.Error())
	}

	allErrors := apiErrors.All()

	if g, w := len(allErrors), 13; g != w {
		t.Fatalf("got %d errors, want %d errors", g, w)
	}
}

func TestErrorsAccessors(t *testing.T) {
	t.Parallel()

	apiErrors := &APIError{}
	err := xml.Unmarshal(errorXML, apiErrors)
	if err != nil {
		t.Fatal("Error unmarshalling: " + err.Error())
	}

	ccObjectErrors := apiErrors.For("Transaction").For("CreditCard")
	if g, w := ccObjectErrors.Object, "CreditCard"; g != w {
		t.Errorf("cc object, got %q, want %q", g, w)
	}

	ccErrors := apiErrors.For("Transaction").For("CreditCard").All()
	if len(ccErrors) != 4 {
		t.Error("Did not get the right credit card errors")
	}

	numberErrors := apiErrors.For("Transaction").For("CreditCard").On("Number")
	if len(numberErrors) != 2 {
		t.Error("Did not get the right number errors")
	}

	customerErrors := apiErrors.For("Transaction").For("Customer").All()
	if len(customerErrors) != 1 {
		t.Error("Did not get the right customer errors")
	}

	lineItemsErrors := apiErrors.For("Transaction").On("LineItems")
	if g, w := len(lineItemsErrors), 1; g != w {
		t.Errorf("line items, got %d, want %d", g, w)
	}

	lineItem1CommodityCodeErrors := apiErrors.For("Transaction").For("LineItems").ForIndex(1).On("CommodityCode")
	if g, w := len(lineItem1CommodityCodeErrors), 1; g != w {
		t.Errorf("line item 1, got %d, want %d", g, w)
	}
	if g, w := lineItem1CommodityCodeErrors[0].Code, "95801"; g != w {
		t.Errorf("line item 1, got %q, want %q", g, w)
	}
	if g, w := lineItem1CommodityCodeErrors[0].Attribute, "CommodityCode"; g != w {
		t.Errorf("line item 1, got %q, want %q", g, w)
	}
	if g, w := lineItem1CommodityCodeErrors[0].Message, "Commodity code is too long."; g != w {
		t.Errorf("line item 1, got %q, want %q", g, w)
	}

	lineItem3Errors := apiErrors.For("Transaction").For("LineItems").ForIndex(3).On("Description")
	if g, w := len(lineItem3Errors), 1; g != w {
		t.Errorf("line item 3, got %d, want %d", g, w)
	}

	baseErrors := apiErrors.For("Transaction").All()
	if g, w := len(baseErrors), 5; g != w {
		t.Errorf("transaction, got %d, want %d", g, w)
	}

	baseBaseErrors := apiErrors.For("Transaction").On("Base")
	if g, w := len(baseBaseErrors), 1; g != w {
		t.Errorf("transaction base, got %d, want %d", g, w)
	}
}

func TestErrorsNameSnakeToCamel(t *testing.T) {
	cases := []struct {
		Snake string
		Camel string
	}{
		{"amount", "Amount"},
		{"index_1", "Index1"},
		{"index_123", "Index123"},
		{"commodity_code", "CommodityCode"},
		{"description", "Description"},
	}

	for _, c := range cases {
		t.Run(c.Snake+"=>"+c.Camel, func(t *testing.T) {
			camel := ErrorNameSnakeToCamel(c.Snake)
			if g, w := camel, c.Camel; g == w {
				t.Logf("got %q, want %q", g, w)
			} else {
				t.Errorf("got %q, want %q", g, w)
			}
		})
	}
}

func TestErrorsNameKebabToCamel(t *testing.T) {
	cases := []struct {
		Kebab string
		Camel string
	}{
		{"amount", "Amount"},
		{"line-items", "LineItems"},
		{"index-1", "Index1"},
		{"index-123", "Index123"},
		{"commodity-code", "CommodityCode"},
		{"description", "Description"},
	}

	for _, c := range cases {
		t.Run(c.Kebab+"=>"+c.Camel, func(t *testing.T) {
			camel := ErrorNameKebabToCamel(c.Kebab)
			if g, w := camel, c.Camel; g == w {
				t.Logf("got %q, want %q", g, w)
			} else {
				t.Errorf("got %q, want %q", g, w)
			}
		})
	}
}

func TestParseKeySignature(t *testing.T) {
	t.Parallel()

	key := NewKey("pubkey", "privkey")

	testCases := []struct {
		Description       string
		SignatureKeyPairs string
		ExpectedSignature string
		ExpectedError     error
	}{
		{
			Description:       "Single signature",
			SignatureKeyPairs: "pubkey|the_signature",
			ExpectedSignature: "the_signature",
		},
		{
			Description:       "Single signature (not matching)",
			SignatureKeyPairs: "pubkey2|the_signature",
			ExpectedError:     SignatureError{Message: "Signature-key pair contains the wrong public key!"},
		},
		{
			Description:       "Multiple signatures (one matching)",
			SignatureKeyPairs: "pubkey|the_signature&pubkey2|the_signature",
			ExpectedSignature: "the_signature",
		},
		{
			Description:       "Invalid signature (no pipe)",
			SignatureKeyPairs: "pubkeythe_signature",
			ExpectedError:     SignatureError{Message: "Signature-key pair does not contain |"},
		},
	}

	for _, tc := range testCases {
		signature, err := key.ParseSignature(tc.SignatureKeyPairs)
		if err != nil && (tc.ExpectedError == nil || err.(SignatureError) != tc.ExpectedError) {
			t.Errorf("Test Case %q with Signature %q encountered error %#v want %#v", tc.Description, tc.SignatureKeyPairs, err, tc.ExpectedError)
		} else if signature != tc.ExpectedSignature {
			t.Errorf("Test Case %q with Signature %q returned %#v want %#v", tc.Description, tc.SignatureKeyPairs, signature, tc.ExpectedSignature)
		}
	}
}

func TestVerifyKeySignature(t *testing.T) {
	t.Parallel()

	key := NewKey("jkq28pcxj4r85dwr", "66062a3876e2dc298f2195f0bf173f5a")

	testCases := []struct {
		Description   string
		Signature     string
		Payload       string
		ExpectedError error
		ExpectedValid bool
	}{
		{
			Description:   "Single signature (valid)",
			Signature:     "jkq28pcxj4r85dwr|4af78bab15cc58195871c636c786716f34cd9711",
			Payload:       "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPG5vdGlm\naWNhdGlvbj4KICA8a2luZD5jaGVjazwva2luZD4KICA8dGltZXN0YW1wIHR5\ncGU9ImRhdGV0aW1lIj4yMDE3LTA0LTI0VDA0OjI1OjEwWjwvdGltZXN0YW1w\nPgogIDxzdWJqZWN0PgogICAgPGNoZWNrIHR5cGU9ImJvb2xlYW4iPnRydWU8\nL2NoZWNrPgogIDwvc3ViamVjdD4KPC9ub3RpZmljYXRpb24+Cg==\n",
			ExpectedValid: true,
		},
		{
			Description:   "Multiple signature (valid)",
			Signature:     "4zn8jg4gdmzyvcyd|dd6390bc9d75985f0cc986d5d5f55fcdb52531cb&cd7jwvrw8jytyfm3|d7fdd777e30a1fd93b58770d7682b577b461cf6f&jkq28pcxj4r85dwr|96dd50905f51f6de1c24790c4d77aa460cb55a3d",
			Payload:       "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPG5vdGlm\naWNhdGlvbj4KICA8a2luZD5jaGVjazwva2luZD4KICA8dGltZXN0YW1wIHR5\ncGU9ImRhdGV0aW1lIj4yMDE3LTA0LTI0VDA0OjUwOjA0WjwvdGltZXN0YW1w\nPgogIDxzdWJqZWN0PgogICAgPGNoZWNrIHR5cGU9ImJvb2xlYW4iPnRydWU8\nL2NoZWNrPgogIDwvc3ViamVjdD4KPC9ub3RpZmljYXRpb24+Cg==\n",
			ExpectedValid: true,
		},
		{
			Description:   "Single signature (invalid)",
			Signature:     "jkq28pcxj4r85dwr|4af78bab15cc58195871c636c786716f34cd9711",
			Payload:       "payloadthatdoesntmatchsignature",
			ExpectedValid: false,
		},
		{
			Description:   "Single signature (unknown public key)",
			Signature:     "cd7jwvrw8jytyfm3|d7fdd777e30a1fd93b58770d7682b577b461cf6f&jkq28pcxj4r85dwr",
			Payload:       "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPG5vdGlm\naWNhdGlvbj4KICA8a2luZD5jaGVjazwva2luZD4KICA8dGltZXN0YW1wIHR5\ncGU9ImRhdGV0aW1lIj4yMDE3LTA0LTI0VDA0OjI1OjEwWjwvdGltZXN0YW1w\nPgogIDxzdWJqZWN0PgogICAgPGNoZWNrIHR5cGU9ImJvb2xlYW4iPnRydWU8\nL2NoZWNrPgogIDwvc3ViamVjdD4KPC9ub3RpZmljYXRpb24+Cg==\n",
			ExpectedError: SignatureError{Message: "Signature-key pair contains the wrong public key!"},
		},
	}

	for _, tc := range testCases {
		valid, err := key.VerifySignature(tc.Signature, tc.Payload)
		if err != nil && (tc.ExpectedError == nil || err.(SignatureError) != tc.ExpectedError) {
			t.Errorf("Test Case %q with Signature %q and Payload %q encountered error %#v want %#v", tc.Description, tc.Signature, tc.Payload, err, tc.ExpectedError)
		} else if valid != tc.ExpectedValid {
			t.Errorf("Test Case %q with Signature %q and Payload %q was valid %#v want %#v", tc.Description, tc.Signature, tc.Payload, valid, tc.ExpectedValid)
		}
	}
}

func TestStripNilElements(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{``, ``},
		{`<element/>`, `<element></element>`},
		{`<element></element>`, `<element></element>`},
		{`<element nil="hi"/>`, `<element nil="hi"></element>`},
		{`<element nil="hi"></element>`, `<element nil="hi"></element>`},
		{`<element nil="true"/>`, ``},
		{`<element nil="true"></element>`, ``},
		{`<root><element><sub>name</sub></element></root>`, `<root><element><sub>name</sub></element></root>`},
		{`<root><element nil="true"><sub>name</sub></element></root>`, `<root></root>`},
		{`<root><element><sub nil="true">name</sub></element></root>`, `<root><element></element></root>`},
	}

	for _, tc := range testCases {
		out, err := StripNilElements([]byte(tc.In))
		if err != nil {
			t.Errorf("StripNilElements(%+v) got err %+v", tc.In, err)
		} else if string(out) != tc.Out {
			t.Errorf("StripNilElements(%+v) got %+v, want %+v", tc.In, string(out), tc.Out)
		} else {
			t.Logf("StripNilElements(%+v) got %q", tc.In, string(out))
		}
	}
}

func TestStripNilElementsErrors(t *testing.T) {
	testCases := []struct {
		In string
	}{
		{`<element>`},
		{`<element><element>`},
		{`<element></sub>`},
		{`<element nil="hi">`},
		{`<element nil="true/>`},
	}

	for _, tc := range testCases {
		_, err := StripNilElements([]byte(tc.In))
		if err == nil {
			t.Errorf("StripNilElements(%+v) got no err", tc.In)
		} else {
			t.Logf("StripNilElements(%+v) got err %+v", tc.In, err)
		}
	}
}

func TestIsTokenNil(t *testing.T) {
	testCases := []struct {
		Token xml.Token
		IsNil bool
	}{
		{xml.Comment{}, false},
		{xml.CharData{}, false},
		{xml.EndElement{}, false},
		{xml.StartElement{}, false},
		{xml.StartElement{Name: xml.Name{Local: "element"}}, false},
	}

	for _, tc := range testCases {
		isNil := IsTokenNil(tc.Token)
		if isNil != tc.IsNil {
			t.Errorf("isTokenNil(%+v) got %+v, want %+v", tc.Token, isNil, tc.IsNil)
		}
	}
}

func TestModificationsXMLEmpty(t *testing.T) {
	t.Parallel()

	m := ModificationsRequest{}
	output, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("got error %#v", err)
	}
	expectedOutput := `<ModificationsRequest></ModificationsRequest>`
	if string(output) != expectedOutput {
		t.Fatalf("got xml %#v, want %#v", string(output), expectedOutput)
	}
}

func TestModificationsXMLMinimalFields(t *testing.T) {
	t.Parallel()

	m := ModificationsRequest{
		Add: []AddModificationRequest{
			{
				InheritedFromID: "1",
			},
			{
				InheritedFromID: "2",
			},
		},
		Update: []UpdateModificationRequest{
			{
				ExistingID: "3",
			},
			{
				ExistingID: "4",
			},
		},
		RemoveExistingIDs: []string{
			"5",
			"6",
		},
	}
	output, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("got error %#v", err)
	}
	expectedOutput := `<ModificationsRequest>
  <add type="array">
    <modification>
      <inherited-from-id>1</inherited-from-id>
    </modification>
    <modification>
      <inherited-from-id>2</inherited-from-id>
    </modification>
  </add>
  <update type="array">
    <modification>
      <existing-id>3</existing-id>
    </modification>
    <modification>
      <existing-id>4</existing-id>
    </modification>
  </update>
  <remove type="array">
    <modification>5</modification>
    <modification>6</modification>
  </remove>
</ModificationsRequest>`
	if string(output) != expectedOutput {
		t.Fatalf("got xml %#v, want %#v", string(output), expectedOutput)
	}
}

func TestModificationsXMLAllFields(t *testing.T) {
	t.Parallel()

	m := ModificationsRequest{
		Add: []AddModificationRequest{
			{
				InheritedFromID: "1",
				ModificationRequest: ModificationRequest{
					Amount:                NewDecimal(100, 2),
					NumberOfBillingCycles: 1,
					Quantity:              1,
					NeverExpires:          true,
				},
			},
			{
				InheritedFromID: "2",
				ModificationRequest: ModificationRequest{
					Amount:                NewDecimal(200, 2),
					NumberOfBillingCycles: 2,
					Quantity:              2,
					NeverExpires:          true,
				},
			},
		},
		Update: []UpdateModificationRequest{
			{
				ExistingID: "3",
				ModificationRequest: ModificationRequest{
					Amount:                NewDecimal(300, 2),
					NumberOfBillingCycles: 3,
					Quantity:              3,
					NeverExpires:          true,
				},
			},
			{
				ExistingID: "4",
				ModificationRequest: ModificationRequest{
					Amount:                NewDecimal(400, 2),
					NumberOfBillingCycles: 4,
					Quantity:              4,
					NeverExpires:          true,
				},
			},
		},
		RemoveExistingIDs: []string{
			"5",
			"6",
		},
	}
	output, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("got error %#v", err)
	}
	expectedOutput := `<ModificationsRequest>
  <add type="array">
    <modification>
      <amount>1.00</amount>
      <number-of-billing-cycles>1</number-of-billing-cycles>
      <quantity>1</quantity>
      <never-expires>true</never-expires>
      <inherited-from-id>1</inherited-from-id>
    </modification>
    <modification>
      <amount>2.00</amount>
      <number-of-billing-cycles>2</number-of-billing-cycles>
      <quantity>2</quantity>
      <never-expires>true</never-expires>
      <inherited-from-id>2</inherited-from-id>
    </modification>
  </add>
  <update type="array">
    <modification>
      <amount>3.00</amount>
      <number-of-billing-cycles>3</number-of-billing-cycles>
      <quantity>3</quantity>
      <never-expires>true</never-expires>
      <existing-id>3</existing-id>
    </modification>
    <modification>
      <amount>4.00</amount>
      <number-of-billing-cycles>4</number-of-billing-cycles>
      <quantity>4</quantity>
      <never-expires>true</never-expires>
      <existing-id>4</existing-id>
    </modification>
  </update>
  <remove type="array">
    <modification>5</modification>
    <modification>6</modification>
  </remove>
</ModificationsRequest>`
	if string(output) != expectedOutput {
		t.Fatalf("got xml %#v, want %#v", string(output), expectedOutput)
	}
}
