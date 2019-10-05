// +build unit

package tests

import (
	"encoding/xml"
	"reflect"
	"testing"
	"time"

	. "github.com/badu/braintree"
)

func TestAndroidPayCardUnmarshalXML(t *testing.T) {
	x := `
	<android-pay-card>
		<source-card-type>mastercard</source-card-type>
		<source-card-last-4>1111</source-card-last-4>
		<virtual-card-type>visa</virtual-card-type>
		<virtual-card-last-4>2222</virtual-card-last-4>
	</android-pay-card>
	`
	var a AndroidPayCard

	if err := xml.Unmarshal([]byte(x), &a); err != nil {
		t.Fatalf("%v", err)
	}

	if g, w := a.SourceCardType, "mastercard"; g != w {
		t.Errorf("SourceCardType got %v, want %v", g, w)
	}
	if g, w := a.SourceCardLast4, "1111"; g != w {
		t.Errorf("SourceCardLast4 got %v, want %v", g, w)
	}
	if g, w := a.VirtualCardType, "visa"; g != w {
		t.Errorf("VirtualCardType got %v, want %v", g, w)
	}
	if g, w := a.VirtualCardLast4, "2222"; g != w {
		t.Errorf("VirtualCardLast4 got %v, want %v", g, w)
	}
	if g, w := a.CardType, "visa"; g != w {
		t.Errorf("CardType got %v, want %v", g, w)
	}
	if g, w := a.Last4, "2222"; g != w {
		t.Errorf("Last4 got %v, want %v", g, w)
	}
}

func TestAndroidPayDetailsUnmarshalXML(t *testing.T) {
	x := `
	<android-pay-card>
		<source-card-type>mastercard</source-card-type>
		<source-card-last-4>1111</source-card-last-4>
		<virtual-card-type>visa</virtual-card-type>
		<virtual-card-last-4>2222</virtual-card-last-4>
	</android-pay-card>
	`
	var a AndroidPayDetail

	if err := xml.Unmarshal([]byte(x), &a); err != nil {
		t.Fatalf("%v", err)
	}

	if g, w := a.SourceCardType, "mastercard"; g != w {
		t.Errorf("SourceCardType got %v, want %v", g, w)
	}
	if g, w := a.SourceCardLast4, "1111"; g != w {
		t.Errorf("SourceCardLast4 got %v, want %v", g, w)
	}
	if g, w := a.VirtualCardType, "visa"; g != w {
		t.Errorf("VirtualCardType got %v, want %v", g, w)
	}
	if g, w := a.VirtualCardLast4, "2222"; g != w {
		t.Errorf("VirtualCardLast4 got %v, want %v", g, w)
	}
	if g, w := a.CardType, "visa"; g != w {
		t.Errorf("CardType got %v, want %v", g, w)
	}
	if g, w := a.Last4, "2222"; g != w {
		t.Errorf("Last4 got %v, want %v", g, w)
	}
}

func TestCustomFieldsMarshalXMLNil(t *testing.T) {
	type object struct {
		Field        string       `xml:"field,omitempty"`
		CustomFields CustomFields `xml:"custom-fields"`
	}
	o := object{}

	output, err := xml.Marshal(o)
	if err != nil {
		t.Fatalf("Error marshaling custom fields: %#v", err)
	}
	xml := string(output)

	if xml != "<object></object>" {
		t.Fatalf("Got %#v, wanted custom fields ommited", xml)
	}
}

func TestCustomFieldsMarshalXMLEmpty(t *testing.T) {
	type object struct {
		Field        string       `xml:"field,omitempty"`
		CustomFields CustomFields `xml:"custom-fields"`
	}
	o := object{CustomFields: CustomFields{}}

	output, err := xml.Marshal(o)
	if err != nil {
		t.Fatalf("Error marshaling custom fields: %#v", err)
	}
	xml := string(output)

	if xml != "<object></object>" {
		t.Fatalf("Got %#v, wanted custom fields ommited", xml)
	}
}

func TestCustomFieldsMarshalXML(t *testing.T) {
	type object struct {
		Field        string       `xml:"field"`
		CustomFields CustomFields `xml:"custom-fields"`
	}
	o := object{
		Field: "1.00",
		CustomFields: CustomFields{
			"custom_field_1": "Custom Value",
		},
	}
	wantedXML := `<object>
  <field>1.00</field>
  <custom-fields>
    <custom-field-1>Custom Value</custom-field-1>
  </custom-fields>
</object>`

	output, err := xml.MarshalIndent(o, "", "  ")
	if err != nil {
		t.Fatalf("Error marshaling custom fields: %#v", err)
	}
	xml := string(output)

	if xml != wantedXML {
		t.Fatalf("Got XML %#v, wanted %#v", xml, wantedXML)
	}
}

func TestCustomFieldsUnmarshalXMLNilEmpty(t *testing.T) {
	type object struct {
		Field        string       `xml:"field"`
		CustomFields CustomFields `xml:"custom-fields"`
	}

	s := `<object>
  <field>1.00</field>
</object>`
	wantedObject := object{
		Field:        "1.00",
		CustomFields: nil,
	}

	o := object{}
	err := xml.Unmarshal([]byte(s), &o)
	if err != nil {
		t.Fatalf("Error marshaling: %#v", err)
	}

	if !reflect.DeepEqual(o, wantedObject) {
		t.Fatalf("Got %#v, wanted %#v", o, wantedObject)
	}
}

func TestCustomFieldsUnmarshalXMLEmpty(t *testing.T) {
	type object struct {
		Field        string       `xml:"field"`
		CustomFields CustomFields `xml:"custom-fields"`
	}

	s := `<object>
  <field>1.00</field>
</object>`
	wantedObject := object{
		Field:        "1.00",
		CustomFields: nil,
	}

	o := object{}
	err := xml.Unmarshal([]byte(s), &o)
	if err != nil {
		t.Fatalf("Error marshaling: %#v", err)
	}

	if !reflect.DeepEqual(o, wantedObject) {
		t.Fatalf("Got %#v, wanted %#v", o, wantedObject)
	}
}

func TestCustomFieldsUnmarshalXMLNil(t *testing.T) {
	type object struct {
		Field        string       `xml:"field"`
		CustomFields CustomFields `xml:"custom-fields"`
	}

	s := `<object>
  <field>1.00</field>
  <custom-fields>
    <custom-field-1>Custom Value One</custom-field-1>
    <custom-field-2>Custom Value Two</custom-field-2>
  </custom-fields>
</object>`
	wantedObject := object{
		Field: "1.00",
		CustomFields: CustomFields{
			"custom_field_1": "Custom Value One",
			"custom_field_2": "Custom Value Two",
		},
	}

	o := object{}
	err := xml.Unmarshal([]byte(s), &o)
	if err != nil {
		t.Fatalf("Error marshaling: %#v", err)
	}

	if !reflect.DeepEqual(o, wantedObject) {
		t.Fatalf("Got %#v, wanted %#v", o, wantedObject)
	}
}

func TestCustomFieldsUnmarshalXML(t *testing.T) {
	type object struct {
		Field        string       `xml:"field"`
		CustomFields CustomFields `xml:"custom-fields"`
	}

	s := `<object>
  <field>1.00</field>
  <custom-fields>
    <custom-field-1>Custom Value One</custom-field-1>
    <custom-field-2>Custom Value Two</custom-field-2>
  </custom-fields>
</object>`
	wantedObject := object{
		Field: "1.00",
		CustomFields: CustomFields{
			"custom_field_1": "Custom Value One",
			"custom_field_2": "Custom Value Two",
		},
	}

	o := object{CustomFields: CustomFields{}}
	err := xml.Unmarshal([]byte(s), &o)
	if err != nil {
		t.Fatalf("Error marshaling: %#v", err)
	}

	if !reflect.DeepEqual(o, wantedObject) {
		t.Fatalf("Got %#v, wanted %#v", o, wantedObject)
	}
}

func TestDateUnmarshalXML(t *testing.T) {
	t.Parallel()

	date := &Date{}

	dateXML := []byte(`<?xml version="1.0" encoding="UTF-8"?><foo>2020-02-09</foo></xml>`)
	if err := xml.Unmarshal(dateXML, date); err != nil {
		t.Fatal(err)
	}

	if date.Format(DateFormat) != "2020-02-09" {
		t.Fatalf("expected 2020-02-09 got %s", date)
	}
}

func TestDateMarshalXML(t *testing.T) {
	t.Parallel()

	date := &Date{Time: time.Date(2020, 2, 9, 0, 0, 0, 0, time.Local)}
	expected := `<Date>2020-02-09</Date>`

	b, err := xml.Marshal(date)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != expected {
		t.Fatalf("expected %s got %s", expected, string(b))
	}
}

func TestDecimalUnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in          []byte
		out         *Decimal
		shouldError bool
	}{
		{[]byte("2.50"), NewDecimal(250, 2), false},
		{[]byte("2"), NewDecimal(2, 0), false},
		{[]byte("0.00"), NewDecimal(0, 2), false},
		{[]byte("-5.504"), NewDecimal(-5504, 3), false},
		{[]byte("0.5"), NewDecimal(5, 1), false},
		{[]byte(".5"), NewDecimal(5, 1), false},
		{[]byte("5.504.98"), NewDecimal(0, 0), true},
		{[]byte("5E6"), NewDecimal(0, 0), true},
	}

	for _, tt := range tests {
		d := &Decimal{}
		err := d.UnmarshalText(tt.in)

		if tt.shouldError {
			if err == nil {
				t.Errorf("expected UnmarshalText(%s) => to error, but it did not", tt.in)
			}
		} else {
			if err != nil {
				t.Errorf("expected UnmarshalText(%s) => to not error, but it did with %s", tt.in, err)
			}
		}

		if !reflect.DeepEqual(d, tt.out) {
			t.Errorf("UnmarshalText(%s) => %+v, want %+v", tt.in, d, tt.out)
		}
	}
}
