package braintree

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

type Search struct {
	fields     []interface{}
	fieldIndex map[string]int
}

func (s *Search) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "search"
	x := struct {
		Fields []interface{}
	}{
		Fields: s.fields,
	}
	return e.EncodeElement(&x, start)
}

type SearchResult struct {
	Page      int
	PageSize  int
	PageCount int
	IDs       []string
}

type TextField struct {
	XMLName    xml.Name
	Is         string `xml:"is,omitempty"`
	IsNot      string `xml:"is-not,omitempty"`
	StartsWith string `xml:"starts-with,omitempty"`
	EndsWith   string `xml:"ends-with,omitempty"`
	Contains   string `xml:"contains,omitempty"`
}

type RangeField struct {
	XMLName xml.Name
	Is      float64 `xml:"is,omitempty"`
	Min     float64 `xml:"min,omitempty"`
	Max     float64 `xml:"max,omitempty"`
}

type TimeField struct {
	XMLName xml.Name
	Is      time.Time
	Min     time.Time
	Max     time.Time
}

func (d TimeField) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = start.Copy()
	start.Name = d.XMLName

	var err error
	err = e.EncodeToken(start)
	if err != nil {
		return err
	}

	err = d.marshalXML(e, "is", d.Is)
	if err != nil {
		return err
	}

	err = d.marshalXML(e, "min", d.Min)
	if err != nil {
		return err
	}

	err = d.marshalXML(e, "max", d.Max)
	if err != nil {
		return err
	}

	err = e.EncodeToken(start.End())
	return err
}

func (d TimeField) marshalXML(e *xml.Encoder, name string, value time.Time) error {
	if value.IsZero() {
		return nil
	}
	const format = "2006-01-02T15:04:05Z"
	start := xml.StartElement{Name: xml.Name{Local: name}}
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "datetime"}}
	return e.EncodeElement(value.UTC().Format(format), start)
}

type MultiField struct {
	XMLName xml.Name
	Type    string   `xml:"type,attr"`
	Items   []string `xml:"item"`
}

func (s *Search) addField(fieldName string, field interface{}) {
	if i, ok := s.fieldIndex[fieldName]; !ok {
		s.fields = append(s.fields, field)
		if s.fieldIndex == nil {
			s.fieldIndex = map[string]int{}
		}
		s.fieldIndex[fieldName] = len(s.fields) - 1
	} else {
		s.fields[i] = field
	}
}

func (s *Search) AddTextField(field string) *TextField {
	f := &TextField{XMLName: xml.Name{Local: field}}
	s.addField(field, f)
	return f
}

func (s *Search) AddRangeField(field string) *RangeField {
	f := &RangeField{XMLName: xml.Name{Local: field}}
	s.addField(field, f)
	return f
}

func (s *Search) AddTimeField(field string) *TimeField {
	f := &TimeField{XMLName: xml.Name{Local: field}}
	s.addField(field, f)
	return f
}

func (s *Search) AddMultiField(field string) *MultiField {
	f := &MultiField{
		XMLName: xml.Name{Local: field},
		Type:    "array",
	}
	s.addField(field, f)
	return f
}

func (s *Search) ShallowCopy() *Search {
	return &Search{
		fields: func() []interface{} {
			a := make([]interface{}, len(s.fields))
			copy(a, s.fields)
			return a
		}(),
		fieldIndex: func() map[string]int {
			m := map[string]int{}
			for f, i := range s.fieldIndex {
				m[f] = i
			}
			return m
		}(),
	}
}

type Decimal struct {
	Unscaled int64
	Scale    int
}

func NewDecimal(unscaled int64, scale int) *Decimal {
	return &Decimal{Unscaled: unscaled, Scale: scale}
}

func (d *Decimal) MarshalText() (text []byte, err error) {
	b := new(bytes.Buffer)
	if d.Scale <= 0 {
		b.WriteString(strconv.FormatInt(d.Unscaled, 10))
		b.WriteString(strings.Repeat("0", -d.Scale))
	} else {
		str := strconv.FormatInt(d.Unscaled, 10)
		if len(str) <= d.Scale {
			str = strings.Repeat("0", (d.Scale+1)-len(str)) + str
		}
		b.WriteString(str[:len(str)-d.Scale])
		b.WriteString(".")
		b.WriteString(str[len(str)-d.Scale:])
	}
	return b.Bytes(), nil
}

func (d *Decimal) UnmarshalText(text []byte) (err error) {
	var (
		str            = string(text)
		unscaled int64 = 0
		scale          = 0
	)

	if str == "" {
		return nil
	}

	if i := strings.Index(str, "."); i != -1 {
		scale = len(str) - i - 1
		str = strings.Replace(str, ".", "", 1)
	}

	if unscaled, err = strconv.ParseInt(str, 10, 64); err != nil {
		return err
	}

	d.Unscaled = unscaled
	d.Scale = scale

	return nil
}

func (d *Decimal) Cmp(y *Decimal) int {
	xUnscaled, yUnscaled := d.Unscaled, y.Unscaled
	xScale, yScale := d.Scale, y.Scale

	for ; xScale > yScale; xScale-- {
		yUnscaled = yUnscaled * 10
	}

	for ; yScale > xScale; yScale-- {
		xUnscaled = xUnscaled * 10
	}

	switch {
	case xUnscaled < yUnscaled:
		return -1
	case xUnscaled > yUnscaled:
		return 1
	default:
		return 0
	}
}

func (d *Decimal) String() string {
	b, err := d.MarshalText()

	if err != nil {
		panic(err)
	}

	return string(b)
}

type Date struct {
	time.Time
}

func (d *Date) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var v string
	err := decoder.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	parse, err := time.Parse(DateFormat, v)
	if err != nil {
		return err
	}

	*d = Date{Time: parse}
	return nil
}

func (d *Date) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.Format(DateFormat), start)
}

type ModificationRequest struct {
	Amount                *Decimal `xml:"amount,omitempty"`
	NumberOfBillingCycles int      `xml:"number-of-billing-cycles,omitempty"`
	Quantity              int      `xml:"quantity,omitempty"`
	NeverExpires          bool     `xml:"never-expires,omitempty"`
}

type AddModificationRequest struct {
	ModificationRequest
	InheritedFromID string `xml:"inherited-from-id,omitempty"`
}

type UpdateModificationRequest struct {
	ModificationRequest
	ExistingID string `xml:"existing-id,omitempty"`
}

type ModificationsRequest struct {
	Add               []AddModificationRequest
	Update            []UpdateModificationRequest
	RemoveExistingIDs []string
}

func (m ModificationsRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type addSchema struct {
		Type          string                   `xml:"type,attr"`
		Modifications []AddModificationRequest `xml:"modification"`
	}
	type updateSchema struct {
		Type          string                      `xml:"type,attr"`
		Modifications []UpdateModificationRequest `xml:"modification"`
	}
	type removeSchema struct {
		Type        string   `xml:"type,attr"`
		ExistingIDs []string `xml:"modification"`
	}
	type schema struct {
		Add    *addSchema    `xml:"add,omitempty"`
		Update *updateSchema `xml:"update,omitempty"`
		Remove *removeSchema `xml:"remove,omitempty"`
	}

	x := schema{}
	if len(m.Add) > 0 {
		x.Add = &addSchema{
			Type:          "array",
			Modifications: m.Add,
		}
	}
	if len(m.Update) > 0 {
		x.Update = &updateSchema{
			Type:          "array",
			Modifications: m.Update,
		}
	}
	if len(m.RemoveExistingIDs) > 0 {
		x.Remove = &removeSchema{
			Type:        "array",
			ExistingIDs: m.RemoveExistingIDs,
		}
	}
	return e.EncodeElement(x, start)
}
