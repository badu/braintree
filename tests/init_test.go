// +build unit integration

package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	. "github.com/badu/braintree"
)

const (
	testCardVisa       = "4111111111111111"
	testCardMastercard = "5555555555554444"
	testCardDiscover   = "6011111111111117"
)

var client *APIClient

var testTimeZone = func() *time.Location {
	tzName := "America/Chicago"

	tz, err := time.LoadLocation(tzName)
	if err != nil {
		panic(fmt.Errorf("error loading time zone location %s: %s", tzName, err))
	}
	return tz
}()

var testMerchantAccountId = "bvps8ckgst59crzs"

// Merchant Account which has AVS and CVV checking turned on.
var avsAndCVVTestMerchantAccountId = "avs-and-cvs"

func testSubMerchantAccount() string {
	acct := MerchantAccount{
		MasterMerchantAccountId: testMerchantAccountId,
		TOSAccepted:             true,
		Id:                      RandomString(),
		Individual: &MerchantAccountPerson{
			FirstName:   "First",
			LastName:    "Last",
			Email:       "harry.belafonte@example.com",
			Phone:       "0000000000",
			DateOfBirth: "1-1-1900",
			Address: &Address{
				StreetAddress:   "222 W Merchandise Mart Plaza",
				ExtendedAddress: "Suite 800",
				Locality:        "Chicago",
				Region:          "IL",
				PostalCode:      "00000",
			},
		},
		FundingOptions: &MerchantAccountFundingOptions{
			Destination: FundingDestMobilePhone,
			MobilePhone: "0000000000",
		},
	}

	merchantAccount, err := client.CreateMerchantAccount(context.Background(), &acct)
	if err != nil {
		panic(fmt.Errorf("error creating test sub merchant account: %s", err))
	}

	return merchantAccount.Id
}

func TestMain(m *testing.M) {
	client = NewWithHttpClient(SandboxURL, "4ngqq224rnk6gvxh", "jkq28pcxj4r85dwr", "66062a3876e2dc298f2195f0bf173f5a", DefaultClient)
	client.Logger = log.New(os.Stderr, "", 0)
	m.Run()
}

var ValidBIN = regexp.MustCompile(`^\d{6}$`).MatchString
var ValidLast4 = regexp.MustCompile(`^\d{4}$`).MatchString
var ValidExpiryMonth = regexp.MustCompile(`^\d{2}$`).MatchString
var ValidExpiryYear = regexp.MustCompile(`^\d{4}$`).MatchString

func IntPtr(i int) *int {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

func RandomInt() int64 {
	var n int64
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	err = binary.Read(buf, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return n
}

func RandomString() string {
	return strconv.FormatInt(RandomInt(), 10)
}

// StringSliceContains checks whether a slice of strings contains
// a given string
func StringSliceContains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		list []string
		s    string
		want bool
	}{
		{
			name: "should return true when list contains s",
			s:    "a",
			list: []string{"a", "b", "c", "d"},
			want: true,
		},
		{
			name: "should return true when list contains s",
			s:    "test3",
			list: []string{"test0", "test1", "test2", "test3", "test4"},
			want: true,
		},
		{
			name: "should return false when list does not contain s",
			s:    "abcd",
			list: []string{"efgh", "ijkl", "mnop", "qrst", "uvwx", "yz"},
			want: false,
		},
		{
			name: "should return false when list does not contain s",
			s:    "a",
			list: []string{"b", "c", "d"},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StringSliceContains(tt.list, tt.s); got != tt.want {
				t.Fatalf("StringSliceContains() => got %v, want %v", got, tt.want)
			}
		})
	}
}
