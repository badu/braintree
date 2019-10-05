// +build integration

package tests

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/badu/braintree"
)

func TestAddOn(t *testing.T) {
	t.Parallel()

	addOns, err := client.ListAddons(context.Background())

	if err != nil {
		t.Fatal(err)
	} else if len(addOns) != 1 {
		t.Fatalf("expected to retrieve one add-on, but retrieved %d", len(addOns))
	}

	addOn := addOns[0]

	if addOn.Id != "test_add_on" {
		t.Fatalf("expected Id to be %s, was %s", "test_add_on", addOn.Id)
	} else if addOn.Amount.Cmp(NewDecimal(1000, 2)) != 0 {
		t.Fatalf("expected Amount to be %s, was %s", NewDecimal(1000, 2), addOn.Amount)
	} else if addOn.Kind != ModificationKindAddOn {
		t.Fatalf("expected Kind to be %s, was %s", ModificationKindAddOn, addOn.Kind)
	} else if addOn.Name != "test_add_on_name" {
		t.Fatalf("expected Name to be %s, was %s", "test_add_on_name", addOn.Name)
	} else if addOn.NeverExpires != true {
		t.Fatalf("expected NeverExpires to be %v, was %v", true, addOn.NeverExpires)
	} else if addOn.Description != "A test add-on" {
		t.Fatalf("expected Description to be %s, was %s", "A test add-on", addOn.Description)
	}
}

func TestAddress(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(
		context.Background(),
		&CustomerRequest{
			FirstName: "Jenna",
			LastName:  "Smith",
		})
	if err != nil {
		t.Fatal(err)
	}
	if customer.Id == "" {
		t.Fatal("invalid customer id")
	}

	addr := &AddressRequest{
		FirstName:          "Jenna",
		LastName:           "Smith",
		Company:            "Braintree",
		StreetAddress:      "1 E Main St",
		ExtendedAddress:    "Suite 403",
		Locality:           "Chicago",
		Region:             "Illinois",
		PostalCode:         "60622",
		CountryCodeAlpha2:  "US",
		CountryCodeAlpha3:  "USA",
		CountryCodeNumeric: "840",
		CountryName:        "United States of America",
	}

	addr2, err := client.CreateAddress(context.Background(), customer.Id, addr)
	if err != nil {
		t.Fatal(err)
	}

	validateAddr(t, addr2, addr, customer)

	addr3 := &AddressRequest{
		FirstName:          "Al",
		LastName:           "Fredidandes",
		Company:            "Paypal",
		StreetAddress:      "1 W Main St",
		ExtendedAddress:    "Suite 402",
		Locality:           "Montreal",
		Region:             "Quebec",
		PostalCode:         "H1A",
		CountryCodeAlpha2:  "CA",
		CountryCodeAlpha3:  "CAN",
		CountryCodeNumeric: "124",
		CountryName:        "Canada",
	}
	addr4, err := client.UpdateAddress(context.Background(), customer.Id, addr2.Id, addr3)
	if err != nil {
		t.Fatal(err)
	}

	validateAddr(t, addr4, addr3, customer)

	err = client.DeleteAddress(context.Background(), customer.Id, addr2.Id)
	if err != nil {
		t.Fatal(err)
	}
}

func validateAddr(t *testing.T, addr *Address, addrRequest *AddressRequest, customer *Customer) {
	if addr.Id == "" {
		t.Fatal("generated id is empty")
	}
	if addr.CustomerId != customer.Id {
		t.Fatal("customer ids do not match")
	}
	if addr.FirstName != addrRequest.FirstName {
		t.Fatal("first names do not match")
	}
	if addr.LastName != addrRequest.LastName {
		t.Fatal("last names do not match")
	}
	if addr.Company != addrRequest.Company {
		t.Fatal("companies do not match")
	}
	if addr.StreetAddress != addrRequest.StreetAddress {
		t.Fatal("street addresses do not match")
	}
	if addr.ExtendedAddress != addrRequest.ExtendedAddress {
		t.Fatal("extended addresses do not match")
	}
	if addr.Locality != addrRequest.Locality {
		t.Fatal("localities do not match")
	}
	if addr.Region != addrRequest.Region {
		t.Fatal("regions do not match")
	}
	if addr.PostalCode != addrRequest.PostalCode {
		t.Fatal("postal codes do not match")
	}
	if addr.CountryCodeAlpha2 != addrRequest.CountryCodeAlpha2 {
		t.Fatal("country alpha2 codes do not match")
	}
	if addr.CountryCodeAlpha3 != addrRequest.CountryCodeAlpha3 {
		t.Fatal("country alpha3 codes do not match")
	}
	if addr.CountryCodeNumeric != addrRequest.CountryCodeNumeric {
		t.Fatal("country numeric codes do not match")
	}
	if addr.CountryName != addrRequest.CountryName {
		t.Fatal("country names do not match")
	}
	if addr.CreatedAt == nil {
		t.Fatal("generated created at is empty")
	}
	if addr.UpdatedAt == nil {
		t.Fatal("generated updated at is empty")
	}
}

func TestClientToken(t *testing.T) {
	t.Parallel()

	token, err := client.GenerateToken(context.Background())
	if err != nil {
		t.Fatalf("failed to generate client token: %s", err)
	}
	if len(token) == 0 {
		t.Fatalf("empty client token!")
	}
}

func TestClientTokenWithCustomer(t *testing.T) {
	t.Parallel()

	customerRequest := &CustomerRequest{FirstName: "Lionel"}

	customer, err := client.CreateCustomer(context.Background(), customerRequest)
	if err != nil {
		t.Error(err)
	}

	customerId := customer.Id

	token, err := client.GenerateWithCustomer(context.Background(), customerId)
	if err != nil {
		t.Error(err)
	} else if len(token) == 0 {
		t.Fatalf("Received empty client token")
	}
}

func TestClientTokenGatewayGenerateWithRequest(t *testing.T) {
	// Getting customer from the API
	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{FirstName: "Brandon"})
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name    string
		req     *TokenRequest
		wantErr bool
	}{
		{
			name: "empty request struct",
			req:  nil,
		},
		{
			name: "request with provided version",
			req:  &TokenRequest{Version: 2},
		},
		{
			name:    "request with non existent customerID",
			req:     &TokenRequest{CustomerID: "///////@@@@@@@"},
			wantErr: true,
		},
		{
			name: "request with customer id",
			req:  &TokenRequest{CustomerID: customer.Id},
		},
		{
			name: "request with merchant id",
			req:  &TokenRequest{MerchantAccountID: testMerchantAccountId},
		},
		{
			name: "request with customer id and merchant id",
			req:  &TokenRequest{CustomerID: customer.Id, MerchantAccountID: testMerchantAccountId},
		},
		{
			name: "request with customer id and merchant id and options verify card not set",
			req: &TokenRequest{
				CustomerID:        customer.Id,
				MerchantAccountID: testMerchantAccountId,
				Options: &Options{
					FailOnDuplicatePaymentMethod: true,
					MakeDefault:                  true,
				},
			},
		},
		{
			name: "request with customer id and merchant id and options verify card true",
			req: &TokenRequest{
				CustomerID:        customer.Id,
				MerchantAccountID: testMerchantAccountId,
				Options: &Options{
					FailOnDuplicatePaymentMethod: true,
					MakeDefault:                  true,
					VerifyCard:                   BoolPtr(true),
				},
			},
		},
		{
			name: "request with customer id and merchant id and options verify card false",
			req: &TokenRequest{
				CustomerID:        customer.Id,
				MerchantAccountID: testMerchantAccountId,
				Options: &Options{
					FailOnDuplicatePaymentMethod: true,
					MakeDefault:                  true,
					VerifyCard:                   BoolPtr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := client
			token, err := g.GenerateWithRequest(context.TODO(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(token) == 0 {
				t.Errorf("empty client token!")
			}
		})
	}
}

func TestCreditCard(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	card, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     cust.Id,
		Number:         testCardVisa,
		ExpirationDate: "05/14",
		CVV:            "100",
		Options: &CreditCardOptions{
			VerifyCard: BoolPtr(true),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if card.Token == "" {
		t.Fatal("invalid token")
	}
	if card.ProductID != "Unknown" { // Unknown appears to be reported from the Braintree Sandbox environment for all cards.
		t.Errorf("got product id %q, want %q", card.ProductID, "Unknown")
	}

	// Update
	card2, err := client.UpdateCard(context.Background(), &CreditCard{
		Token:          card.Token,
		Number:         testCardMastercard,
		ExpirationDate: "05/14",
		CVV:            "100",
		Options: &CreditCardOptions{
			VerifyCard: BoolPtr(true),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if card2.Token != card.Token {
		t.Fatal("tokens do not match")
	}
	if card2.CardType != "MasterCard" {
		t.Fatal("card type does not match")
	}

	// Delete
	err = client.DeleteCard(context.Background(), card2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreditCardFailedAutoVerification(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateCard(context.Background(), &CreditCard{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
	})
	if err == nil {
		t.Fatal("Got no error, want error")
	}
	if g, w := err.(*APIError).ErrorMessage, "Do Not Honor"; g != w {
		t.Fatalf("Got error %q, want error %q", g, w)
	}
}

func TestCreditCardForceNotVerified(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateCard(context.Background(), &CreditCard{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
		Options: &CreditCardOptions{
			VerifyCard: BoolPtr(false),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateCreditCardWithExpirationMonthAndYear(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	card, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:      customer.Id,
		Number:          testCardVisa,
		ExpirationMonth: "05",
		ExpirationYear:  "2014",
		CVV:             "100",
	})

	if err != nil {
		t.Fatal(err)
	}
	if card.Token == "" {
		t.Fatal("invalid token")
	}
}

func TestCreateCreditCardInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := client.CreateCard(context.Background(), &CreditCard{
		Number:         testCardVisa,
		ExpirationDate: "05/14",
	})

	if err == nil {
		t.Fatal("expected to get error creating card because of required fields, but did not")
	}
}

func TestFindCreditCard(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	card, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     customer.Id,
		Number:         testCardVisa,
		ExpirationDate: "05/14",
		CVV:            "100",
		Options: &CreditCardOptions{
			VerifyCard: BoolPtr(true),
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if card.Token == "" {
		t.Fatal("invalid token")
	}

	card2, err := client.FindCard(context.Background(), card.Token)

	if err != nil {
		t.Fatal(err)
	}
	if card2.Token != card.Token {
		t.Fatal("tokens do not match")
	}
}

func TestFindCreditCardBadData(t *testing.T) {
	t.Parallel()

	_, err := client.FindCard(context.Background(), "invalid_token")

	if err == nil {
		t.Fatal("expected to get error because the token is invalid")
	}
}

func TestSaveCreditCardWithVenmoSDKPaymentMethodCode(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	card, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:                customer.Id,
		VenmoSDKPaymentMethodCode: "stub-" + testCardVisa,
	})
	if err != nil {
		t.Fatal(err)
	}
	if card.VenmoSDK {
		t.Fatal("venmo card marked")
	}
}

func TestSaveCreditCardWithVenmoSDKSession(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	card, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     customer.Id,
		Number:         testCardVisa,
		ExpirationDate: "05/14",
		Options: &CreditCardOptions{
			VenmoSDKSession: "stub-session",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if card.VenmoSDK {
		t.Fatal("venmo card marked")
	}
}

func TestGetExpiringBetweenCards(t *testing.T) {
	now := time.Now()

	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	card1, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     customer.Id,
		Number:         testCardVisa,
		ExpirationDate: now.AddDate(0, -2, 0).Format("01/2006"),
	})
	if err != nil {
		t.Fatal(err)
	}

	card2, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     customer.Id,
		Number:         testCardVisa,
		ExpirationDate: now.Format("01/2006"),
	})
	if err != nil {
		t.Fatal(err)
	}

	card3, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:     customer.Id,
		Number:         testCardVisa,
		ExpirationDate: now.AddDate(0, 2, 0).Format("01/2006"),
	})
	if err != nil {
		t.Fatal(err)
	}

	fromDate := now.AddDate(0, -1, 0)
	toDate := now.AddDate(0, 1, 0)

	expiringCards := map[string]bool{}
	results, err := client.ExpiringBetween(context.Background(), fromDate, toDate)
	if err != nil {
		t.Fatal(err)
	}
	for page := 1; page <= results.PageCount; page++ {
		results.Page = page
		resultz, err := client.ExpiringBetweenPaged(context.Background(), fromDate, toDate, results)
		if err != nil {
			t.Fatal(err)
		}
		for _, card := range resultz.CreditCards {
			expiringCards[card.Token] = true
		}
	}
	if expiringCards[card1.Token] {
		t.Fatalf("expiringCards contains card1 (%s), it shouldn't be returned in expiring cards results", card1.Token)
	}
	if !expiringCards[card2.Token] {
		t.Fatalf("expiringCards does not contain card2 (%s), it should be returned in expiring cards results", card2.Token)
	}
	if expiringCards[card3.Token] {
		t.Fatalf("expiringCards contains card3 (%s), it shouldn't be returned in expiring cards results", card3.Token)
	}
	results.Page = 0
	_, err = client.ExpiringBetweenPaged(context.Background(), fromDate, toDate, results)
	if err == nil || !strings.Contains(err.Error(), "page 0 out of bounds") {
		t.Errorf("requesting page 0 should result in out of bounds error, but got %#v", err)
	}
	results.Page = results.PageCount + 1
	_, err = client.ExpiringBetweenPaged(context.Background(), fromDate, toDate, results)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("page %d out of bounds", results.PageCount+1)) {
		t.Errorf("requesting page %d should result in out of bounds error, but got %v", results.PageCount+1, err)
	}
}

func TestCustomerAndroidPayCard(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	nonce := FakeNonceAndroidPay

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if customerFound.AndroidPayCards == nil || len(customerFound.AndroidPayCards.AndroidPayCard) != 1 {
		t.Fatalf("Customer %#v expected to have one AndroidPayCard", customerFound)
	}
	customerCard := customerFound.AndroidPayCards.AndroidPayCard[0]
	if customerCard.Token != paymentMethod.Token ||
		customerCard.Default != paymentMethod.Default ||
		customerCard.ImageURL != paymentMethod.ImageURL ||
		customerCard.CustomerId != paymentMethod.CustomerId {
		t.Fatalf("Got Customer %#v AndroidPayCard %#v, want %#v", customerFound, customerCard, paymentMethod)
	}
}

func TestCustomerApplePayCard(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceApplePayVisa,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if customerFound.ApplePayCards == nil || len(customerFound.ApplePayCards.Cards) != 1 {
		t.Fatalf("Customer %#v expected to have one ApplePayCard", customerFound)
	}
	customerCard := customerFound.ApplePayCards.Cards[0]
	if customerCard.Token != paymentMethod.Token ||
		customerCard.Default != paymentMethod.Default ||
		customerCard.ImageURL != paymentMethod.ImageURL ||
		customerCard.CustomerId != paymentMethod.CustomerId {
		t.Fatalf("Got Customer %#v AndroidPayCard %#v, want %#v", customerFound, customerCard, paymentMethod)
	}
}

const (
	FakeNonceTransactable                       = "fake-valid-nonce"
	FakeNonceConsumed                           = "fake-consumed-nonce"
	FakeNoncePayPalOneTimePayment               = "fake-paypal-one-time-nonce"
	FakeNoncePayPalFuturePayment                = "fake-paypal-future-nonce"
	FakeNonceApplePayVisa                       = "fake-apple-pay-visa-nonce"
	FakeNonceApplePayMastercard                 = "fake-apple-pay-mastercard-nonce"
	FakeNonceApplePayAmex                       = "fake-apple-pay-amex-nonce"
	FakeNonceAbstractTransactable               = "fake-abstract-transactable-nonce"
	FakeNoncePayPalBillingAgreement             = "fake-paypal-billing-agreement-nonce"
	FakeNonceEurope                             = "fake-europe-bank-account-nonce"
	FakeNonceCoinbase                           = "fake-coinbase-nonce"
	FakeNonceAndroidPay                         = "fake-android-pay-nonce"
	FakeNonceAndroidPayDiscover                 = "fake-android-pay-discover-nonce"
	FakeNonceAndroidPayVisa                     = "fake-android-pay-visa-nonce"
	FakeNonceAndroidPayMasterCard               = "fake-android-pay-mastercard-nonce"
	FakeNonceAndroidPayAmEx                     = "fake-android-pay-amex-nonce"
	FakeNonceAmexExpressCheckout                = "fake-amex-express-checkout-nonce"
	FakeNonceVenmoAccount                       = "fake-venmo-account-nonce"
	FakeNonceTransactableVisa                   = "fake-valid-visa-nonce"
	FakeNonceTransactableAmEx                   = "fake-valid-amex-nonce"
	FakeNonceTransactableMasterCard             = "fake-valid-mastercard-nonce"
	FakeNonceTransactableDiscover               = "fake-valid-discover-nonce"
	FakeNonceTransactableJCB                    = "fake-valid-jcb-nonce"
	FakeNonceTransactableMaestro                = "fake-valid-maestro-nonce"
	FakeNonceTransactableDinersClub             = "fake-valid-dinersclub-nonce"
	FakeNonceTransactablePrepaid                = "fake-valid-prepaid-nonce"
	FakeNonceTransactableCommercial             = "fake-valid-commercial-nonce"
	FakeNonceTransactableDurbinRegulated        = "fake-valid-durbin-regulated-nonce"
	FakeNonceTransactableHealthcare             = "fake-valid-healthcare-nonce"
	FakeNonceTransactableDebit                  = "fake-valid-debit-nonce"
	FakeNonceTransactablePayroll                = "fake-valid-payroll-nonce"
	FakeNonceTransactableNoIndicators           = "fake-valid-no-indicators-nonce"
	FakeNonceTransactableUnknownIndicators      = "fake-valid-unknown-indicators-nonce"
	FakeNonceTransactableCountryOfIssuanceUSA   = "fake-valid-country-of-issuance-usa-nonce"
	FakeNonceTransactableCountryOfIssuanceCAD   = "fake-valid-country-of-issuance-cad-nonce"
	FakeNonceTransactableIssuingBankNetworkOnly = "fake-valid-issuing-bank-network-only-nonce"
	FakeNonceProcessorDeclinedVisa              = "fake-processor-declined-visa-nonce"
	FakeNonceProcessorDeclinedMasterCard        = "fake-processor-declined-mastercard-nonce"
	FakeNonceProcessorDeclinedAmEx              = "fake-processor-declined-amex-nonce"
	FakeNonceProcessorDeclinedDiscover          = "fake-processor-declined-discover-nonce"
	FakeNonceProcessorFailureJCB                = "fake-processor-failure-jcb-nonce"
	FakeNonceLuhnInvalid                        = "fake-luhn-invalid-nonce"
	FakeNoncePayPalFuturePaymentRefreshToken    = "fake-paypal-future-refresh-token-nonce"
	FakeNonceSEPA                               = "fake-sepa-bank-account-nonce"
	FakeNonceMasterpassAmEx                     = "fake-masterpass-amex-nonce"
	FakeNonceMasterpassDiscover                 = "fake-masterpass-discover-nonce"
	FakeNonceMasterpassMaestro                  = "fake-masterpass-maestro-nonce"
	FakeNonceMasterpassMasterCard               = "fake-masterpass-mastercard-nonce"
	FakeNonceMasterpassVisa                     = "fake-masterpass-visa-nonce"
	FakeNonceGatewayRejectedFraud               = "fake-gateway-rejected-fraud-nonce"
	FakeNonceVisaCheckoutVisa                   = "fake-visa-checkout-visa-nonce"
	FakeNonceVisaCheckoutAmEx                   = "fake-visa-checkout-amex-nonce"
	FakeNonceVisaCheckoutMasterCard             = "fake-visa-checkout-mastercard-nonce"
	FakeNonceVisaCheckoutDiscover               = "fake-visa-checkout-discover-nonce"
)

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestCustomer(t *testing.T) {
	t.Parallel()

	oc := &CustomerRequest{
		FirstName: "Bogdan",
		LastName:  "Dinu",
		Company:   "Appetize",
		Email:     "bogdan.dinu@example.com",
		Phone:     "312.555.1234",
		Fax:       "614.555.5678",
		Website:   "http://www.example.com",
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "200",
			Options: &CreditCardOptions{
				VerifyCard: BoolPtr(true),
			},
		},
	}

	// Create with errors
	_, err := client.CreateCustomer(context.Background(), oc)
	if err == nil {
		t.Fatal("Did not receive error when creating invalid customer")
	}

	// Create
	oc.CreditCard.CVV = ""
	oc.CreditCard.Options = nil
	customer, err := client.CreateCustomer(context.Background(), oc)

	if err != nil {
		t.Fatal(err)
	}
	if customer.Id == "" {
		t.Fatal("invalid customer id")
	}
	if card := customer.DefaultCreditCard(); card == nil {
		t.Fatal("invalid credit card")
	}
	if card := customer.DefaultCreditCard(); card.Token == "" {
		t.Fatal("invalid token")
	}
	if customer.CreatedAt == nil {
		t.Fatal("generated created at is empty")
	}
	if customer.UpdatedAt == nil {
		t.Fatal("generated updated at is empty")
	}

	// Update
	unique := RandomString()
	newFirstName := "John" + unique
	c2, err := client.UpdateCustomer(context.Background(), &CustomerRequest{
		ID:        customer.Id,
		FirstName: newFirstName,
	})

	if err != nil {
		t.Fatal(err)
	}
	if c2.FirstName != newFirstName {
		t.Fatal("first name not changed")
	}

	// Find
	c3, err := client.FindCustomer(context.Background(), customer.Id)

	if err != nil {
		t.Fatal(err)
	}
	if c3.Id != customer.Id {
		t.Fatal("ids do not match")
	}

	// Search
	query := new(Search)
	f := query.AddTextField("first-name")
	f.Is = newFirstName

	idsResult, err := client.SearchCustomersByIDs(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	idsResult.Page = 1
	searchResult, err := client.SearchCustomer(context.Background(), query, idsResult)

	if err != nil {
		t.Fatal(err)
	}
	if len(searchResult.Customers) == 0 {
		t.Fatal("could not search for a customer")
	}
	if id := searchResult.Customers[0].Id; id != customer.Id {
		t.Fatalf("id from search does not match: got %s, wanted %s", id, customer.Id)
	}

	// Delete
	err = client.DeleteCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	// Test customer 404
	c4, err := client.FindCustomer(context.Background(), customer.Id)
	if err == nil {
		t.Fatal("should return 404")
	}
	if err.Error() != "Not Found (404)" {
		t.Fatal(err)
	}
	if apiErr, ok := err.(IAPIError); !(ok && apiErr.StatusCode() == http.StatusNotFound) {
		t.Fatal(err)
	}
	if c4 != nil {
		t.Fatal(c4)
	}
}

func TestCustomerSearch(t *testing.T) {
	t.Parallel()

	const customersCount = 51
	customerIDs := map[string]bool{}
	prefix := "PaginationTest-" + RandomString()
	for i := 0; i < customersCount; i++ {
		unique := RandomString()
		tx, err := client.CreateCustomer(context.Background(), &CustomerRequest{
			FirstName: "John",
			LastName:  "Smith",
			Company:   prefix + unique,
		})
		if err != nil {
			t.Fatal(err)
		}
		customerIDs[tx.Id] = true
	}

	query := new(Search)
	query.AddTextField("company").StartsWith = prefix

	results, err := client.SearchCustomersByIDs(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.IDs) != customersCount {
		t.Fatalf("results.IDs = %v, want %v", len(results.IDs), customersCount)
	}

	for page := 1; page <= results.PageCount; page++ {
		results.Page = page
		resultz, err := client.SearchCustomer(context.Background(), query, results)
		if err != nil {
			t.Fatal(err)
		}
		for _, cs := range resultz.Customers {
			if company := cs.Company; !strings.HasPrefix(company, prefix) {
				t.Fatalf("cs.Company = %q, want prefix of %q", company, prefix)
			}
			if customerIDs[cs.Id] {
				delete(customerIDs, cs.Id)
			} else {
				t.Fatalf("cs.Id = %q, not expected", cs.Id)
			}
		}
	}

	if len(customerIDs) > 0 {
		t.Fatalf("customers not returned = %v", customerIDs)
	}
	results.Page = 0
	_, err = client.SearchCustomer(context.Background(), query, results)

	if err == nil || !strings.Contains(err.Error(), "page 0 out of bounds") {
		t.Errorf("requesting page 0 should result in out of bounds error, but got %#v", err)
	}
	results.Page = results.PageCount + 1
	_, err = client.SearchCustomer(context.Background(), query, results)

	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("page %d out of bounds", results.PageCount+1)) {
		t.Errorf("requesting page %d should result in out of bounds error, but got %v", results.PageCount+1, err)
	}

}

func TestCustomerWithCustomFields(t *testing.T) {
	t.Parallel()

	customFields := map[string]string{
		"custom_field_1": "custom value",
	}

	c := &CustomerRequest{
		CustomFields: customFields,
	}

	customer, err := client.CreateCustomer(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}

	if x := map[string]string(customer.CustomFields); !reflect.DeepEqual(x, customFields) {
		t.Fatalf("Returned custom fields doesn't match input, got %q, want %q", x, customFields)
	}

	customer, err = client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if x := map[string]string(customer.CustomFields); !reflect.DeepEqual(x, customFields) {
		t.Fatalf("Returned custom fields doesn't match input, got %q, want %q", x, customFields)
	}
}

func TestCustomerPaymentMethods(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	paymentMethod1, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNoncePayPalBillingAgreement,
	})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod2, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedPaymentMethods := []PaymentMethod{
		*paymentMethod2,
		*paymentMethod1,
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(customerFound.PaymentMethods(), expectedPaymentMethods) {
		t.Fatalf("Got Customer %#v PaymentMethods %#v, want %#v", customerFound, customerFound.PaymentMethods(), expectedPaymentMethods)
	}
}

func TestCustomerDefaultPaymentMethod(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	defaultPaymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNoncePayPalBillingAgreement,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(customerFound.DefaultPaymentMethod(), defaultPaymentMethod) {
		t.Fatalf("Got Customer %#v DefaultPaymentMethod %#v, want %#v", customerFound, customerFound.DefaultPaymentMethod(), defaultPaymentMethod)
	}
}

func TestCustomerPaymentMethodNonce(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{PaymentMethodNonce: FakeNonceTransactable})
	if err != nil {
		t.Fatal(err)
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if len(customer.PaymentMethods()) != 1 {
		t.Fatalf("Customer %#v has %#v payment method(s), want 1 payment method", customerFound, len(customer.PaymentMethods()))
	}
}

func TestCustomerAddresses(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{
		FirstName: "Jenna",
		LastName:  "Smith",
	})
	if err != nil {
		t.Fatal(err)
	}
	if customer.Id == "" {
		t.Fatal("invalid customer id")
	}

	addrReqs := []*AddressRequest{
		&AddressRequest{
			FirstName:          "Jenna",
			LastName:           "Smith",
			Company:            "Braintree",
			StreetAddress:      "1 E Main St",
			ExtendedAddress:    "Suite 403",
			Locality:           "Chicago",
			Region:             "Illinois",
			PostalCode:         "60622",
			CountryCodeAlpha2:  "US",
			CountryCodeAlpha3:  "USA",
			CountryCodeNumeric: "840",
			CountryName:        "United States of America",
		},
		{
			FirstName:          "Bob",
			LastName:           "Rob",
			Company:            "Paypal",
			StreetAddress:      "1 W Main St",
			ExtendedAddress:    "Suite 402",
			Locality:           "Boston",
			Region:             "Massachusetts",
			PostalCode:         "02140",
			CountryCodeAlpha2:  "US",
			CountryCodeAlpha3:  "USA",
			CountryCodeNumeric: "840",
			CountryName:        "United States of America",
		},
	}

	for _, addrReq := range addrReqs {
		_, err = client.CreateAddress(context.Background(), customer.Id, addrReq)
		if err != nil {
			t.Fatal(err)
		}
	}

	customerWithAddrs, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if customerWithAddrs.Addresses == nil || len(customerWithAddrs.Addresses.Address) != 2 {
		t.Fatal("wrong number of addresses returned")
	}

	for _, addr := range customerWithAddrs.Addresses.Address {
		if addr.Id == "" {
			t.Fatal("generated id is empty")
		}

		var addrReq *AddressRequest
		for _, ar := range addrReqs {
			if ar.PostalCode == addr.PostalCode {
				addrReq = ar
				break
			}
		}

		if addrReq == nil {
			t.Fatal("did not return sent address")
		}

		if addr.CustomerId != customer.Id {
			t.Errorf("got customer id %s, want %s", addr.CustomerId, customer.Id)
		}
		if addr.FirstName != addrReq.FirstName {
			t.Errorf("got first name %s, want %s", addr.FirstName, addrReq.FirstName)
		}
		if addr.LastName != addrReq.LastName {
			t.Errorf("got last name %s, want %s", addr.LastName, addrReq.LastName)
		}
		if addr.Company != addrReq.Company {
			t.Errorf("got company %s, want %s", addr.Company, addrReq.Company)
		}
		if addr.StreetAddress != addrReq.StreetAddress {
			t.Errorf("got street address %s, want %s", addr.StreetAddress, addrReq.StreetAddress)
		}
		if addr.ExtendedAddress != addrReq.ExtendedAddress {
			t.Errorf("got extended address %s, want %s", addr.ExtendedAddress, addrReq.ExtendedAddress)
		}
		if addr.Locality != addrReq.Locality {
			t.Errorf("got locality %s, want %s", addr.Locality, addrReq.Locality)
		}
		if addr.Region != addrReq.Region {
			t.Errorf("got region %s, want %s", addr.Region, addrReq.Region)
		}
		if addr.CountryCodeAlpha2 != addrReq.CountryCodeAlpha2 {
			t.Errorf("got country code alpha 2 %s, want %s", addr.CountryCodeAlpha2, addrReq.CountryCodeAlpha2)
		}
		if addr.CountryCodeAlpha3 != addrReq.CountryCodeAlpha3 {
			t.Errorf("got country code alpha 3 %s, want %s", addr.CountryCodeAlpha3, addrReq.CountryCodeAlpha3)
		}
		if addr.CountryCodeNumeric != addrReq.CountryCodeNumeric {
			t.Errorf("got country code numeric %s, want %s", addr.CountryCodeNumeric, addrReq.CountryCodeNumeric)
		}
		if addr.CountryName != addrReq.CountryName {
			t.Errorf("got country name %s, want %s", addr.CountryName, addrReq.CountryName)
		}
		if addr.CreatedAt == nil {
			t.Error("got created at nil, want a value")
		}
		if addr.UpdatedAt == nil {
			t.Error("got updated at nil, want a value")
		}
	}

	err = client.DeleteCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCustomerPayPalAccount(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	nonce := FakeNoncePayPalFuturePayment

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}
	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if customerFound.PayPalAccounts == nil || len(customerFound.PayPalAccounts.Accounts) != 1 {
		t.Fatalf("Customer %#v expected to have one PayPalAccount", customerFound)
	}
	customerCard := customerFound.PayPalAccounts.Accounts[0]
	if customerCard.Token != paymentMethod.Token ||
		customerCard.Default != paymentMethod.Default ||
		customerCard.ImageURL != paymentMethod.ImageURL ||
		customerCard.CustomerId != paymentMethod.CustomerId {
		t.Fatalf("Got Customer %#v PayPalAccount %#v, want %#v", customerFound, customerCard, paymentMethod)
	}
}

func TestCustomerVenmoAccount(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	nonce := FakeNonceVenmoAccount

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerFound, err := client.FindCustomer(context.Background(), customer.Id)
	if err != nil {
		t.Fatal(err)
	}

	if customerFound.VenmoAccounts == nil || len(customerFound.VenmoAccounts.Accounts) != 1 {
		t.Fatalf("Customer %#v expected to have one VenmoAccount", customerFound)
	}
	customerCard := customerFound.VenmoAccounts.Accounts[0]
	if customerCard.Token != paymentMethod.Token ||
		customerCard.Default != paymentMethod.Default ||
		customerCard.ImageURL != paymentMethod.ImageURL ||
		customerCard.CustomerId != paymentMethod.CustomerId {
		t.Fatalf("Got Customer %#v VenmoAccount %#v, want %#v", customerFound, customerCard, paymentMethod)
	}
}

func TestDisbursementTransactions(t *testing.T) {
	t.Parallel()

	d := Disbursement{
		TransactionIds: []string{"dskdmb"},
	}

	result, err := d.Transactions(context.Background(), client)

	if err != nil {
		t.Fatal(err)
	}

	if result.TotalItems != 1 {
		t.Fatal(result)
	}

	txn := result.Transactions[0]
	if txn.Id != "dskdmb" {
		t.Fatal(txn.Id)
	}
}

func TestDiscounts(t *testing.T) {
	t.Parallel()

	discounts, err := client.AllDiscounts(context.Background())

	if err != nil {
		t.Error(err)
	} else if len(discounts) != 1 {
		t.Fatalf("expected to retrieve 1 discount, retrieved %d", len(discounts))
	}

	discount := discounts[0]

	if discount.Id != "test_discount" {
		t.Fatalf("expected Id to be %s, was %s", "test_discount", discount.Id)
	} else if discount.Amount.Cmp(NewDecimal(1000, 2)) != 0 {
		t.Fatalf("expected Amount to be %s, was %s", NewDecimal(1000, 2), discount.Amount)
	} else if discount.Kind != ModificationKindDiscount {
		t.Fatalf("expected Kind to be %s, was %s", ModificationKindDiscount, discount.Kind)
	} else if discount.Name != "test_discount_name" {
		t.Fatalf("expected Name to be %s, was %s", "test_discount_name", discount.Name)
	} else if discount.NeverExpires != true {
		t.Fatalf("expected NeverExpires to be %t, was %t", true, discount.NeverExpires)
	} else if discount.CurrentBillingCycle != 0 {
		t.Fatalf("expected current billing cycle to be %d, was %d", 0, discount.CurrentBillingCycle)
	} else if discount.Description != "A test discount" {
		t.Fatalf("expected Description to be %s, was %s", "A test discount", discount.Description)
	}
}

func TestDisputeFinalize(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(100, 2),
		CreditCard: &CreditCard{
			Number:         "4023898493988028",
			ExpirationDate: "12/" + time.Now().Format("2006"),
		},
		Options: &TxOpts{
			SubmitForSettlement: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to create disputed transaction: %v", err)
	}

	tx, err = client.FindTransaction(context.Background(), tx.Id)
	if err != nil {
		t.Fatalf("failed to find disputed transaction: %v", err)
	}

	if len(tx.Disputes) != 1 {
		t.Fatalf("got Tx with %d disputes, want 1", len(tx.Disputes))
	}

	dispute := tx.Disputes[0]

	if dispute.AmountDisputed.Cmp(NewDecimal(100, 2)) != 0 {
		t.Errorf("got Dispute AmountDisputed %s, want %s", dispute.AmountDisputed, "1.00")
	}

	err = client.Finalize(context.Background(), dispute.ID)

	if err != nil {
		t.Fatalf("failed to finalize dispute: %v", err)
	}
}

func TestDisputeAccept(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(100, 2),
		CreditCard: &CreditCard{
			Number:         "4023898493988028",
			ExpirationDate: "12/" + time.Now().Format("2006"),
		},
		Options: &TxOpts{
			SubmitForSettlement: true,
		},
	})

	if err != nil {
		t.Fatalf("failed to create disputed transaction: %v", err)
	}

	tx, err = client.FindTransaction(context.Background(), tx.Id)
	if err != nil {
		t.Fatalf("failed to find disputed transaction: %v", err)
	}

	if len(tx.Disputes) != 1 {
		t.Fatalf("transaction has %d disputes, want 1", len(tx.Disputes))
	}

	dispute := tx.Disputes[0]

	err = client.Accept(context.Background(), dispute.ID)
	if err != nil {
		t.Fatalf("failed to accept dispute: %v", err)
	}

	dispute, err = client.FindDispute(context.Background(), dispute.ID)
	if err != nil {
		t.Fatalf("failed to find dispute: %v", err)
	}

	if dispute.Status != DisputeStatusAccepted {
		t.Fatalf("got Dispute Status %q, want %q", dispute.Status, DisputeStatusAccepted)
	}
}

func TestDisputeTextEvidence(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(100, 2),
		CreditCard: &CreditCard{
			Number:         "4023898493988028",
			ExpirationDate: "12/" + time.Now().Format("2006"),
		},
		Options: &TxOpts{
			SubmitForSettlement: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to create disputed transaction: %v", err)
	}

	tx, err = client.FindTransaction(context.Background(), tx.Id)
	if err != nil {
		t.Fatalf("failed to find disputed transaction: %v", err)
	}

	if len(tx.Disputes) != 1 {
		t.Fatalf("got Tx with %d disputes, want 1", len(tx.Disputes))
	}

	dispute := tx.Disputes[0]

	textEvidence, err := client.AddTextEvidence(context.Background(), dispute.ID, &DisputeTextEvidenceRequest{
		Content:  "some evidence",
		Category: DisputeDeviceName,
	})
	if err != nil {
		t.Fatalf("failed to add text evidence: %v", err)
	}

	if textEvidence.ID == "" {
		t.Fatal("text evidence can not have empty id")
	}

	err = client.RemoveEvidence(context.Background(), dispute.ID, textEvidence.ID)
	if err != nil {
		t.Fatalf("failed to remove evidence: %v", err)
	}

	err = client.Finalize(context.Background(), dispute.ID)
	if err != nil {
		t.Fatalf("failed to finalize dispute: %v", err)
	}
}

func TestMerchantAccountCreate(t *testing.T) {
	t.Parallel()

	acct := MerchantAccount{
		MasterMerchantAccountId: testMerchantAccountId,
		TOSAccepted:             true,
		Id:                      RandomString(),
		Individual: &MerchantAccountPerson{
			FirstName:   "Don",
			LastName:    "Johnson",
			Email:       "bogdan.dinu@example.com",
			Phone:       "5556789012",
			DateOfBirth: "1-1-1989",
			Address: &Address{
				StreetAddress:   "1 E Main St",
				ExtendedAddress: "Suite 404",
				Locality:        "Chicago",
				Region:          "IL",
				PostalCode:      "60622",
			},
		},
		FundingOptions: &MerchantAccountFundingOptions{
			Destination: FundingDestMobilePhone,
			MobilePhone: "5552344567",
		},
	}

	merchantAccount, err := client.CreateMerchantAccount(context.Background(), &acct)

	if err != nil {
		t.Fatal(err)
	}

	if merchantAccount.Id == "" {
		t.Fatal("invalid merchant account id")
	}

	ma2, err := client.FindMerchantAccount(context.Background(), merchantAccount.Id)

	if err != nil {
		t.Fatal(err)
	}

	if ma2.Id != merchantAccount.Id {
		t.Fatal("ids do not match")
	}

}

func TestPaymentMethod(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	// Create using credit card
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNonceTransactableVisa,
	})
	if err != nil {
		t.Fatal(err)
	}

	if paymentMethod.CustomerId != cust.Id {
		t.Errorf("Got paymentMethod customer Id %#v, want %#v", paymentMethod.CustomerId, cust.Id)
	}
	if paymentMethod.Token == "" {
		t.Errorf("Got paymentMethod token %#v, want a value", paymentMethod.Token)
	}

	// Update using different credit card
	rand.Seed(time.Now().UTC().UnixNano())
	token := fmt.Sprintf("btgo_test_token_%d", rand.Int()+1)
	paymentMethod, err = client.UpdatePayMethod(context.Background(), paymentMethod.Token, &PaymentMethodRequest{
		PaymentMethodNonce: FakeNonceTransactableMasterCard,
		Token:              token,
	})
	if err != nil {
		t.Fatal(err)
	}

	if paymentMethod.Token != token {
		t.Errorf("Got paymentMethod token %#v, want %#v", paymentMethod.Token, token)
	}

	// Updating with different payment method type should fail
	if _, err = client.UpdatePayMethod(context.Background(), token, &PaymentMethodRequest{PaymentMethodNonce: FakeNoncePayPalBillingAgreement}); err == nil {
		t.Errorf("Updating with a different payment method type should have failed")
	}

	// Find credit card
	paymentMethod, err = client.FindPayMethod(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}

	if paymentMethod.CustomerId != cust.Id {
		t.Errorf("Got paymentMethod customer Id %#v, want %#v", paymentMethod.CustomerId, cust.Id)
	}
	if paymentMethod.Token != token {
		t.Errorf("Got paymentMethod token %#v, want %#v", paymentMethod.Token, token)
	}

	// Delete credit card
	if err := client.DeletePayMethod(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	// Create using PayPal
	paymentMethod, err = client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNoncePayPalBillingAgreement,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Find PayPal
	_, err = client.FindPayMethod(context.Background(), paymentMethod.Token)
	if err != nil {
		t.Fatal(err)
	}

	// Updating a PayPal account with a different payment method nonce of any kind should fail
	if _, err = client.UpdatePayMethod(context.Background(), paymentMethod.Token, &PaymentMethodRequest{PaymentMethodNonce: FakeNoncePayPalOneTimePayment}); err == nil {
		t.Errorf("Updating a PayPal account with a different nonce should have failed")
	}

	// Delete PayPal
	if err := client.DeletePayMethod(context.Background(), paymentMethod.Token); err != nil {
		t.Fatal(err)
	}

	// Cleanup
	if err := client.DeleteCustomer(context.Background(), cust.Id); err != nil {
		t.Fatal(err)
	}
}

func TestPaymentMethodFailedAutoVerification(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
	})
	if err == nil {
		t.Fatal("Got no error, want error")
	}
	if g, w := err.(*APIError).ErrorMessage, "Do Not Honor"; g != w {
		t.Fatalf("Got error %q, want error %q", g, w)
	}
}

func TestPaymentMethodForceNotVerified(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         cust.Id,
		PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
		Options: &PaymentMethodRequestOptions{
			VerifyCard: BoolPtr(false),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

}

func TestPaymentMethodNonce(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactableVisa,
	})
	if err != nil {
		t.Fatal(err)
	}

	paymentMethodNonce, err := client.CreatePaymentMethodNonce(context.Background(), paymentMethod.Token)
	if err != nil {
		t.Fatal(err)
	}

	if paymentMethodNonce.Type != "CreditCard" {
		t.Errorf("nonce type got %q, want %q", paymentMethodNonce.Type, "CreditCard")
	}

	paymentMethodNonceFound, err := client.FindPaymentMethodNonce(context.Background(), paymentMethodNonce.Nonce)
	if err != nil {
		t.Fatal(err)
	}

	if paymentMethodNonceFound.Type != "CreditCard" {
		t.Errorf("found nonce type got %q, want %q", paymentMethodNonceFound.Type, "CreditCard")
	}
	if paymentMethodNonceFound.Nonce != paymentMethodNonce.Nonce {
		t.Errorf("found nonce got %q, want %q", paymentMethodNonceFound.Nonce, paymentMethodNonce.Nonce)
	}
}

func TestPayPalAccount(t *testing.T) {
	t.Parallel()

	cust, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	nonce := FakeNoncePayPalBillingAgreement

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         cust.Id,
		PaymentMethodNonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Find
	paypalAccount, err := client.SearchPaypalAccount(context.Background(), paymentMethod.Token)
	if err != nil {
		t.Fatal(err)
	}

	if paypalAccount.Token == "" {
		t.Fatal("invalid token")
	}

	// Update
	paypalAccount2, err := client.UpdatePaypalAccount(context.Background(), &PayPalAccount{
		Token: paypalAccount.Token,
		Email: "new-email@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	if paypalAccount2.Token != paypalAccount.Token {
		t.Fatal("tokens do not match")
	}
	if paypalAccount2.Email != "new-email@example.com" {
		t.Fatal("paypalAccount email does not match")
	}

	// Delete
	err = client.DeletePaypalAccount(context.Background(), paypalAccount2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindPayPalAccountBadData(t *testing.T) {
	t.Parallel()

	_, err := client.SearchPaypalAccount(context.Background(), "invalid_token")

	if err == nil {
		t.Fatal("expected to get error because the token is invalid")
	}
}

func TestPlan(t *testing.T) {
	t.Parallel()

	plans, err := client.ListPlans(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) == 0 {
		t.Fatal(plans)
	}

	var plan *Plan
	for _, p := range plans {
		if p.Id == "test_plan" {
			plan = p
			break
		}
	}

	if plan == nil {
		t.Fatal("plan not found")
	}
	if x := plan.Id; x != "test_plan" {
		t.Fatal(x)
	}
	if x := plan.MerchantId; x == "" {
		t.Fatal(x)
	}
	if x := plan.BillingFrequency; x == nil || *x != 1 {
		t.Fatal(x)
	}
	if x := plan.CurrencyISOCode; x != "USD" {
		t.Fatal(x)
	}
	if x := plan.Description; x != "test_plan_desc" {
		t.Fatal(x)
	}
	if x := plan.Name; x != "test_plan_name" {
		t.Fatal(x)
	}
	if x := plan.NumberOfBillingCycles; x == nil || *x != 2 {
		t.Fatal(x)
	}
	if x := plan.Price; x.Cmp(NewDecimal(1000, 2)) != 0 {
		t.Fatal(x)
	}
	if x := plan.TrialDuration; x == nil || *x != 14 {
		t.Fatal(x)
	}
	if x := plan.TrialDurationUnit; x != "day" {
		t.Fatal(x)
	}
	if x := plan.TrialPeriod; !x {
		t.Fatal(x)
	}
	if x := plan.CreatedAt; x == nil {
		t.Fatal(x)
	}
	if x := plan.UpdatedAt; x == nil {
		t.Fatal(x)
	}

	// Add Ons
	if len(plan.AddOns.AddOns) == 0 {
		t.Fatal(plan.AddOns)
	}
	addOn := plan.AddOns.AddOns[0]

	if addOn.Id != "test_add_on" {
		t.Fatalf("expected Id to be %s, was %s", "test_add_on", addOn.Id)
	} else if addOn.Amount.Cmp(NewDecimal(1000, 2)) != 0 {
		t.Fatalf("expected Amount to be %s, was %s", NewDecimal(1000, 2), addOn.Amount)
	} else if addOn.Kind != ModificationKindAddOn {
		t.Fatalf("expected Kind to be %s, was %s", ModificationKindAddOn, addOn.Kind)
	} else if addOn.Name != "test_add_on_name" {
		t.Fatalf("expected Name to be %s, was %s", "test_add_on_name", addOn.Name)
	} else if addOn.NeverExpires != true {
		t.Fatalf("expected NeverExpires to be %v, was %v", true, addOn.NeverExpires)
	} else if addOn.Description != "A test add-on" {
		t.Fatalf("expected Description to be %s, was %s", "A test add-on", addOn.Description)
	} else if addOn.NumberOfBillingCycles != 0 {
		t.Fatalf("expected NumberOfBillingCycles to be %d, was %d", 0, addOn.NumberOfBillingCycles)
	}

	// Discounts
	if len(plan.Discounts.Discounts) == 0 {
		t.Fatal(plan.Discounts)
	}
	discount := plan.Discounts.Discounts[0]

	if discount.Id != "test_discount" {
		t.Fatalf("expected Id to be %s, was %s", "test_discount", discount.Id)
	} else if discount.Amount.Cmp(NewDecimal(1000, 2)) != 0 {
		t.Fatalf("expected Amount to be %s, was %s", NewDecimal(1000, 2), discount.Amount)
	} else if discount.Kind != ModificationKindDiscount {
		t.Fatalf("expected Kind to be %s, was %s", ModificationKindDiscount, discount.Kind)
	} else if discount.Name != "test_discount_name" {
		t.Fatalf("expected Name to be %s, was %s", "test_discount_name", discount.Name)
	} else if discount.NeverExpires != true {
		t.Fatalf("expected NeverExpires to be %v, was %v", true, discount.NeverExpires)
	} else if discount.Description != "A test discount" {
		t.Fatalf("expected Description to be %s, was %s", "A test discount", discount.Description)
	} else if discount.NumberOfBillingCycles != 0 {
		t.Fatalf("expected NumberOfBillingCycles to be %d, was %d", 0, discount.NumberOfBillingCycles)
	}

	// Find
	plan2, err := client.FindPlan(context.Background(), "test_plan_2")
	if err != nil {
		t.Fatal(err)
	}
	if plan2.Id != "test_plan_2" {
		t.Fatal(plan2)
	}
	if len(plan2.AddOns.AddOns) != 0 {
		t.Fatal(plan2.AddOns)
	}
	if len(plan2.Discounts.Discounts) != 0 {
		t.Fatal(plan2.Discounts)
	}
}

func TestSettlementBatch(t *testing.T) {
	t.Parallel()

	// Create a new transaction
	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1000, 2),
		PaymentMethodNonce: FakeNonceTransactableJCB,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	// Submit for settlement
	tx, err = client.SubmitForSettlement(context.Background(), tx.Id, tx.Amount)
	if err != nil {
		t.Fatal(err)
	}
	if x := tx.Status; x != StatusSubmittedForSettlement {
		t.Fatal(x)
	}

	// Settle
	tx, err = client.SandboxSettle(context.Background(), tx.Id)
	if err != nil {
		t.Fatal(err)
	}
	if x := tx.Status; x != StatusSettled {
		t.Fatal(x)
	}

	// Generate Settlement Batch Summary which will include new transaction
	date := tx.SettlementBatchId[:10]
	summary, err := client.GenerateSettlement(context.Background(), &Settlement{Date: date})
	if err != nil {
		t.Fatalf("unable to get settlement batch: %s", err)
	}

	var found bool
	for _, r := range summary.Records.Type {
		if r.MerchantAccountId == tx.MerchantAccountId && r.CardType == tx.CreditCard.CardType && r.Count > 0 && r.Kind == "sale" {
			found = true
		}
	}

	if !found {
		t.Fatalf("Tx %s created but no record in the settlement batch for it's merchant account and card type.", tx.Id)
	}
}

func TestSubscriptionSimple(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	sub, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if len(sub.StatusEvents) != 1 {
		t.Fatalf("expected one status event, got %d", len(sub.StatusEvents))
	}
	wantBalance := NewDecimal(0, 2)
	wantPrice := NewDecimal(1000, 2)
	event := sub.StatusEvents[0]
	if event.Status != SubscriptionStatusActive {
		t.Fatalf("expected status of status history event to be active, was %s", event.Status)
	}
	if event.CurrencyISOCode != "USD" {
		t.Fatalf("expected currency iso code of status history event to be USD, was %s", event.CurrencyISOCode)
	}
	if event.Balance.Cmp(wantBalance) != 0 {
		t.Fatalf("expected balance of status history event to be 0, was %s", event.Balance)
	}
	if event.Price.Cmp(wantPrice) != 0 {
		t.Fatalf("expected price of status history event to be 10, was %s", event.Price)
	}

	// Update
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	newId := fmt.Sprintf("%X", b[:])
	sub2, err := client.UpdateSubscription(context.Background(), sub.Id, &SubscriptionRequest{
		Id:     newId,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != newId {
		t.Fatalf("expected subscription ID to change to %s but is %s", newId, sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}
	if len(sub2.StatusEvents) != 2 {
		t.Fatalf("expected two status events, got %d", len(sub2.StatusEvents))
	}
	for _, event := range sub2.StatusEvents {
		if event.Status != SubscriptionStatusActive {
			t.Fatalf("expected status of status history event to be active, was %s", event.Status)
		}
		if event.CurrencyISOCode != "USD" {
			t.Fatalf("expected currency iso code of status history event to be USD, was %s", event.CurrencyISOCode)
		}
		if event.Balance.Cmp(wantBalance) != 0 {
			t.Fatalf("expected balance of status history event to be 0, was %s", event.Balance)
		}
		if event.Price.Cmp(wantPrice) != 0 {
			t.Fatalf("expected price of status history event to be 10, was %s", event.Price)
		}
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub2.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub2.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	_, err = client.CancelSubscription(context.Background(), sub2.Id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionAllFieldsWithBillingDayOfMonth(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken:    paymentMethod.Token,
		PlanId:                "test_plan",
		MerchantAccountId:     testMerchantAccountId,
		BillingDayOfMonth:     IntPtr(15),
		NumberOfBillingCycles: IntPtr(2),
		Price:                 NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != "15" {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, "15")
	}
	if x := sub1.NeverExpires; x {
		t.Fatalf("got never expires %#v, want false", x)
	}
	if x := sub1.NumberOfBillingCycles; x == nil || *x != 2 {
		t.Fatalf("got number billing cycles %#v, want 2", x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if x := sub1.Status; x != SubscriptionStatusPending && x != SubscriptionStatusActive {
		t.Fatalf("got status %#v, want Pending or Active (it will be active if todays date matches the billing day of month)", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}
	if sub1.CreatedAt == nil {
		t.Fatal("expected createdAt to not be nil")
	}
	if sub1.UpdatedAt == nil {
		t.Fatal("expected updatedAt to not be nil")
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	sub4, err := client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if x := sub4.Status; x != SubscriptionStatusCanceled {
		t.Fatalf("got status %#v, want Canceled", x)
	}
}

func TestSubscriptionAllFieldsWithBillingDayOfMonthNeverExpires(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
		MerchantAccountId:  testMerchantAccountId,
		BillingDayOfMonth:  IntPtr(15),
		NeverExpires:       BoolPtr(true),
		Price:              NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != "15" {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, "15")
	}
	if x := sub1.NeverExpires; !x {
		t.Fatalf("got never expires %#v, want true", x)
	}
	if x := sub1.NumberOfBillingCycles; x != nil {
		t.Fatalf("got number billing cycles %#v, didn't want", x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if x := sub1.Status; x != SubscriptionStatusPending && x != SubscriptionStatusActive {
		t.Fatalf("got status %#v, want Pending or Active (it will be active if todays date matches the billing day of month)", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		Id:     sub1.Id,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	sub4, err := client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if x := sub4.Status; x != SubscriptionStatusCanceled {
		t.Fatalf("got status %#v, want Canceled", x)
	}
}

func TestSubscriptionAllFieldsWithFirstBillingDate(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	firstBillingDate := fmt.Sprintf("%d-12-31", time.Now().Year())
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken:    paymentMethod.Token,
		PlanId:                "test_plan",
		MerchantAccountId:     testMerchantAccountId,
		FirstBillingDate:      firstBillingDate,
		NumberOfBillingCycles: IntPtr(2),
		Price:                 NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != "31" {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, "31")
	}
	if sub1.FirstBillingDate != firstBillingDate {
		t.Fatalf("got first billing date %#v, want %#v", sub1.FirstBillingDate, firstBillingDate)
	}
	if x := sub1.NeverExpires; x {
		t.Fatalf("got never expires %#v, want false", x)
	}
	if x := sub1.NumberOfBillingCycles; x == nil {
		t.Fatalf("got number billing cycles nil, want 2")
	} else if *x != 2 {
		t.Fatalf("got number billing cycles %#v, want 2", *x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if x := sub1.Status; x != SubscriptionStatusPending {
		t.Fatalf("got status %#v, want Pending", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		Id:     sub1.Id,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	sub4, err := client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if x := sub4.Status; x != SubscriptionStatusCanceled {
		t.Fatalf("got status %#v, want Canceled", x)
	}
}

func TestSubscriptionAllFieldsWithFirstBillingDateNeverExpires(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	firstBillingDate := fmt.Sprintf("%d-12-31", time.Now().Year())
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
		MerchantAccountId:  testMerchantAccountId,
		FirstBillingDate:   firstBillingDate,
		NeverExpires:       BoolPtr(true),
		Price:              NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != "31" {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, "31")
	}
	if sub1.FirstBillingDate != firstBillingDate {
		t.Fatalf("got first billing date %#v, want %#v", sub1.FirstBillingDate, firstBillingDate)
	}
	if x := sub1.NeverExpires; !x {
		t.Fatalf("got never expires %#v, want true", x)
	}
	if x := sub1.NumberOfBillingCycles; x != nil {
		t.Fatalf("got number billing cycles %#v, didn't want", x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if x := sub1.Status; x != SubscriptionStatusPending {
		t.Fatalf("got status %#v, want Pending", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		Id:     sub1.Id,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	sub4, err := client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if x := sub4.Status; x != SubscriptionStatusCanceled {
		t.Fatalf("got status %#v, want Canceled", x)
	}
}

func TestSubscriptionAllFieldsWithTrialPeriod(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	firstBillingDate := time.Now().In(testTimeZone).AddDate(0, 0, 7)
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken:    paymentMethod.Token,
		PlanId:                "test_plan",
		MerchantAccountId:     testMerchantAccountId,
		TrialPeriod:           BoolPtr(true),
		TrialDuration:         "7",
		TrialDurationUnit:     Day,
		NumberOfBillingCycles: IntPtr(2),
		Price:                 NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != fmt.Sprintf("%d", firstBillingDate.Day()) {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, firstBillingDate.Day())
	}
	if sub1.FirstBillingDate != firstBillingDate.Format(DateFormat) {
		t.Fatalf("got first billing date %#v, want %#v", sub1.FirstBillingDate, firstBillingDate)
	}
	if x := sub1.NeverExpires; x {
		t.Fatalf("got never expires %#v, want false", x)
	}
	if x := sub1.NumberOfBillingCycles; x == nil || *x != 2 {
		t.Fatalf("got number billing cycles %#v, want 2", x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; !x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if sub1.TrialDuration != "7" {
		t.Fatalf("got trial duration %#v, want 7", sub1.TrialDuration)
	}
	if sub1.TrialDurationUnit != Day {
		t.Fatalf("got trial duration unit %#v, want day", sub1.TrialDurationUnit)
	}
	if x := sub1.Status; x != SubscriptionStatusActive {
		t.Fatalf("got status %#v, want Active", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		Id:     sub1.Id,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	_, err = client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionAllFieldsWithTrialPeriodNeverExpires(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	firstBillingDate := time.Now().In(testTimeZone).AddDate(0, 0, 7)
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
		MerchantAccountId:  testMerchantAccountId,
		TrialPeriod:        BoolPtr(true),
		TrialDuration:      "7",
		TrialDurationUnit:  Day,
		NeverExpires:       BoolPtr(true),
		Price:              NewDecimal(100, 2),
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub1.Id == "" {
		t.Fatal("invalid subscription id")
	}
	if sub1.BillingDayOfMonth != fmt.Sprintf("%d", firstBillingDate.Day()) {
		t.Fatalf("got billing day of month %#v, want %#v", sub1.BillingDayOfMonth, firstBillingDate.Day())
	}
	if sub1.FirstBillingDate != firstBillingDate.Format(DateFormat) {
		t.Fatalf("got first billing date %#v, want %#v", sub1.FirstBillingDate, firstBillingDate)
	}
	if x := sub1.NeverExpires; !x {
		t.Fatalf("got never expires %#v, want true", x)
	}
	if x := sub1.NumberOfBillingCycles; x != nil {
		t.Fatalf("got number billing cycles %#v, didn't want", x)
	}
	if x := sub1.Price; x == nil || x.Scale != 2 || x.Unscaled != 100 {
		t.Fatalf("got price %#v, want 1.00", x)
	}
	if x := sub1.TrialPeriod; !x {
		t.Fatalf("got trial period %#v, want false", x)
	}
	if sub1.TrialDuration != "7" {
		t.Fatalf("got trial duration %#v, want 7", sub1.TrialDuration)
	}
	if sub1.TrialDurationUnit != Day {
		t.Fatalf("got trial duration unit %#v, want day", sub1.TrialDurationUnit)
	}
	if x := sub1.Status; x != SubscriptionStatusActive {
		t.Fatalf("got status %#v, want Active", x)
	}
	if x := sub1.Descriptor.Name; x != "Company Name*Product 1" {
		t.Fatalf("got descriptor name %#v, want Company Name*Product 1", x)
	}
	if x := sub1.Descriptor.Phone; x != "0000000000" {
		t.Fatalf("got descriptor phone %#v, want 0000000000", x)
	}
	if x := sub1.Descriptor.URL; x != "example.com" {
		t.Fatalf("got descriptor url %#v, want example.com", x)
	}

	// Update
	sub2, err := client.UpdateSubscription(context.Background(), sub1.Id, &SubscriptionRequest{
		Id:     sub1.Id,
		PlanId: "test_plan_2",
		Options: &SubscriptionOpts{
			ProrateCharges:                       true,
			RevertSubscriptionOnProrationFailure: true,
			StartImmediately:                     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub1.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}

	// Find
	sub3, err := client.FindSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub1.Id {
		t.Fatal(sub3.Id)
	}

	// Cancel
	_, err = client.CancelSubscription(context.Background(), sub1.Id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionModifications(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	sub, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan_2",
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub.Id == "" {
		t.Fatal("invalid subscription id")
	}

	// Add AddOn
	sub2, err := client.UpdateSubscription(context.Background(), sub.Id, &SubscriptionRequest{
		Id: sub.Id,
		AddOns: &ModificationsRequest{
			Add: []AddModificationRequest{
				{
					InheritedFromID: "test_add_on",
					ModificationRequest: ModificationRequest{
						Amount:       NewDecimal(300, 2),
						Quantity:     1,
						NeverExpires: true,
					},
				},
			},
		},
		Discounts: &ModificationsRequest{
			Add: []AddModificationRequest{
				{
					InheritedFromID: "test_discount",
					ModificationRequest: ModificationRequest{
						Amount:                NewDecimal(100, 2),
						Quantity:              1,
						NumberOfBillingCycles: 2,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}
	if x := sub2.AddOns.AddOns; len(x) != 1 {
		t.Fatalf("got %d add ons, want 1 add on", len(x))
	}
	if x := sub2.AddOns.AddOns[0].Amount; x.String() != NewDecimal(300, 2).String() {
		t.Fatalf("got %v add on, want 3.00 add on", x)
	}
	if x := sub2.Discounts.Discounts; len(x) != 1 {
		t.Fatalf("got %d discounts, want 1 discount", len(x))
	}
	if x := sub2.Discounts.Discounts[0].Amount; x.String() != NewDecimal(100, 2).String() {
		t.Fatalf("got %v discount, want 1.00 discount", x)
	}
	if x := sub2.Discounts.Discounts[0].NumberOfBillingCycles; x != 2 {
		t.Fatalf("got %v number of billing cycles on discount, want 2 billing cycles", x)
	}
	if x := sub2.Discounts.Discounts[0].CurrentBillingCycle; x != 0 {
		t.Fatalf("got current billing cycle of %d on discount, want 0", x)
	}

	// Update AddOn
	sub3, err := client.UpdateSubscription(context.Background(), sub.Id, &SubscriptionRequest{
		Id: sub.Id,
		AddOns: &ModificationsRequest{
			Update: []UpdateModificationRequest{
				{
					ExistingID: "test_add_on",
					ModificationRequest: ModificationRequest{
						Amount: NewDecimal(150, 2),
					},
				},
			},
		},
		Discounts: &ModificationsRequest{
			RemoveExistingIDs: []string{
				"test_discount",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub3.Id != sub.Id {
		t.Fatal(sub3.Id)
	}
	if x := sub3.PlanId; x != "test_plan_2" {
		t.Fatal(x)
	}
	if x := sub3.AddOns.AddOns; len(x) != 1 {
		t.Fatalf("got %d add ons, want 1 add on", len(x))
	}
	if x := sub3.AddOns.AddOns[0].Amount; x.String() != NewDecimal(150, 2).String() {
		t.Fatalf("got %v add on, want 1.50 add on", x)
	}
	if x := sub3.Discounts.Discounts; len(x) != 0 {
		t.Fatalf("got %d discounts, want 0 discounts", len(x))
	}

	// Cancel
	_, err = client.CancelSubscription(context.Background(), sub3.Id)
	if err != nil {
		t.Fatal(err)
	}
}

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestSubscriptionTransactions(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	sub, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
		Options: &SubscriptionOpts{
			StartImmediately: true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if sub.Id == "" {
		t.Fatal("invalid subscription id")
	}

	// Find
	sub2, err := client.FindSubscription(context.Background(), sub.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sub2.Id != sub.Id {
		t.Fatal(sub2.Id)
	}
	if x := sub2.PlanId; x != "test_plan" {
		t.Fatal(x)
	}
	if len(sub2.Transactions.Transaction) < 1 {
		t.Fatalf("Expected transactions slice not to be empty")
	}
	if x := sub2.Transactions.Transaction[0].PlanId; x != "test_plan" {
		t.Fatal(x)
	}
	if x := sub2.Transactions.Transaction[0].SubscriptionId; x != sub.Id {
		t.Fatal(x)
	}
	if x := sub2.Transactions.Transaction[0].SubscriptionDetails.BillingPeriodStartDate; x != sub.BillingPeriodStartDate {
		t.Fatal(x)
	}
	if x := sub2.Transactions.Transaction[0].SubscriptionDetails.BillingPeriodEndDate; x != sub.BillingPeriodEndDate {
		t.Fatal(x)
	}

	// Cancel
	_, err = client.CancelSubscription(context.Background(), sub2.Id)
	if err != nil {
		t.Fatal(err)
	}
}

// It is not possible to successfully retry a charge without manually creating
// a subcription with a card that will fail, waiting a day for it to be billed
// and fail which will cause the subscription to enter the PastDue status. This
// test instead attempts to retry a charge that is not PastDue and ensures the
// only errors returned is the status.
// Ref: https://developers.braintreepayments.com/guides/recurring-billing/overview#past-due
func TestSubscriptionRetryCharge(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	verifyCard := false
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
		Options: &PaymentMethodRequestOptions{
			VerifyCard: &verifyCard,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create Subscription
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
		MerchantAccountId:  testMerchantAccountId,
		Price:              NewDecimal(100, 2),
		Options: &SubscriptionOpts{
			StartImmediately: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, err = client.CancelSubscription(context.Background(), sub1.Id)
		if err != nil {
			t.Error(err)
		}
	}()

	// Retry Charge
	err = client.RetryCharge(context.Background(), &SubscriptionTransactionRequest{
		SubscriptionID: sub1.Id,
		Amount:         NewDecimal(10, 2),
		Options: &SubscriptionTransactionOptionsRequest{
			SubmitForSettlement: true,
		},
	})
	if err == nil {
		t.Fatalf("Retry charge did not error, want error indicating Subscription status must be Past Due in order to retry.")
	}
	btErr, ok := err.(*APIError)
	if !ok {
		t.Fatal(err)
	}
	validationErrs := btErr.All()
	if len(validationErrs) != 1 {
		t.Fatalf("got %d validation errors, want 1, validation errors: %#v", len(validationErrs), validationErrs)
	}
	wantValidationErr := ValidationError{
		Code:      "91531",
		Attribute: "Base",
		Message:   "Subscription status must be Past Due in order to retry.",
	}
	if validationErrs[0] != wantValidationErr {
		t.Errorf("got validation error %#v, want %#v", validationErrs[0], wantValidationErr)
	}
}

func TestSubscriptionSearchIDs(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
	})
	sub2, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
	})
	sub3, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan_2",
	})

	query := &Search{}
	f1 := query.AddTimeField("created-at")
	f1.Max = time.Now()
	f1.Min = time.Now().AddDate(0, 0, -1)
	f2 := query.AddTextField("plan-id")
	f2.Is = "test_plan"

	result, err := client.SearchSubscriptions(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if !StringSliceContains(result.IDs, sub1.Id) {
		t.Errorf("expected result.IDs to include %v", sub1.Id)
	}
	if !StringSliceContains(result.IDs, sub2.Id) {
		t.Errorf("expected result.IDs to include %v", sub2.Id)
	}
	if StringSliceContains(result.IDs, sub3.Id) {
		t.Errorf("expected result.Ids to not include %v", sub3.Id)
	}
}

func TestSubscriptionSearch(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}
	sub1, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
	})
	sub2, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan",
	})
	sub3, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
		PaymentMethodToken: paymentMethod.Token,
		PlanId:             "test_plan_2",
	})

	query := &Search{}
	f1 := query.AddTimeField("created-at")
	f1.Max = time.Now()
	f1.Min = time.Now().Add(-10 * time.Minute)
	f2 := query.AddTextField("plan-id")
	f2.Is = "test_plan"

	searchResult, err := client.SearchSubscriptions(context.Background(), query)
	if err != nil {
		t.Fatalf("error : %v", err)
	}

	pageSize := searchResult.PageSize
	ids := searchResult.IDs

	endOffset := pageSize
	if endOffset > len(ids) {
		endOffset = len(ids)
	}

	firstPageQuery := query.ShallowCopy()
	firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
	firstPageSubscriptions, err := client.FetchSubscriptions(context.Background(), firstPageQuery)

	result := &SubscriptionSearchResult{
		TotalItems:        len(ids),
		TotalIDs:          ids,
		CurrentPageNumber: 1,
		PageSize:          pageSize,
		Subscriptions:     firstPageSubscriptions,
	}
	if err != nil {
		t.Fatal(err)
	}

	if result.CurrentPageNumber != 1 {
		t.Errorf("expected page number to be 1, got %v", result.CurrentPageNumber)
	}
	if !StringSliceContains(result.TotalIDs, sub1.Id) {
		t.Errorf("expected subscription ids to contain %v", sub1.Id)
	}
	if !StringSliceContains(result.TotalIDs, sub2.Id) {
		t.Errorf("expected subscription ids to contain %v", sub2.Id)
	}
	if StringSliceContains(result.TotalIDs, sub3.Id) {
		t.Errorf("expected subscription ids to not contain %v", sub3.Id)
	}
}

func TestSubscriptionSearchNext(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	const subscriptionCount = 51
	expectedIDs := map[string]bool{}
	for i := 0; i < subscriptionCount; i++ {
		sub, err := client.CreateSubscription(context.Background(), &SubscriptionRequest{
			PaymentMethodToken: paymentMethod.Token,
			PlanId:             "test_plan_2",
		})
		if err != nil {
			t.Fatal(err)
		}
		expectedIDs[sub.Id] = true
	}

	query := &Search{}
	f1 := query.AddTimeField("created-at")
	f1.Max = time.Now()
	f1.Min = time.Now().Add(-10 * time.Minute)
	f2 := query.AddTextField("plan-id")
	f2.Is = "test_plan_2"

	searchResult, err := client.SearchSubscriptions(context.Background(), query)
	if err != nil {
		t.Fatalf("error : %v", err)
	}

	pageSize := searchResult.PageSize
	ids := searchResult.IDs

	endOffset := pageSize
	if endOffset > len(ids) {
		endOffset = len(ids)
	}

	firstPageQuery := query.ShallowCopy()
	firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
	firstPageSubscriptions, err := client.FetchSubscriptions(context.Background(), firstPageQuery)

	result := &SubscriptionSearchResult{
		TotalItems:        len(ids),
		TotalIDs:          ids,
		CurrentPageNumber: 1,
		PageSize:          pageSize,
		Subscriptions:     firstPageSubscriptions,
	}
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalItems < subscriptionCount {
		t.Errorf("result.TotalItems = %v, want it to be more than %v", result.TotalItems, subscriptionCount)
	}

	for {
		for _, sub := range result.Subscriptions {
			if expectedIDs[sub.Id] {
				delete(expectedIDs, sub.Id)
			}
		}

		startOffset := result.CurrentPageNumber * result.PageSize
		endOffset := startOffset + result.PageSize
		if endOffset > len(result.TotalIDs) {
			endOffset = len(result.TotalIDs)
		}
		if startOffset >= endOffset {

		} else {

			nextPageQuery := query.ShallowCopy()
			nextPageQuery.AddMultiField("ids").Items = result.TotalIDs[startOffset:endOffset]
			nextPageSubscriptions, err := client.FetchSubscriptions(context.Background(), nextPageQuery)
			if err != nil {
				t.Fatalf("error : %v", err)
			}
			result = &SubscriptionSearchResult{
				TotalItems:        result.TotalItems,
				TotalIDs:          result.TotalIDs,
				CurrentPageNumber: result.CurrentPageNumber + 1,
				PageSize:          result.PageSize,
				Subscriptions:     nextPageSubscriptions,
			}
		}

		if result == nil {
			break
		}
	}

	if len(expectedIDs) > 0 {
		t.Fatalf("subscriptions not returned = %v", expectedIDs)
	}
}

func TestTransaction3DSRequiredGatewayRejected(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(1007, 2)

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	cc, err := client.CreateCard(context.Background(), &CreditCard{
		CustomerId:      customer.Id,
		Number:          testCardVisa,
		ExpirationYear:  "2020",
		ExpirationMonth: "01",
	})
	if err != nil {
		t.Fatal(err)
	}

	nonce, err := client.CreatePaymentMethodNonce(context.Background(), cc.Token)
	if err != nil {
		t.Fatal(err)
	}
	if nonce.ThreeDSecureInfo != nil {
		t.Fatalf("Nonce 3DS Info present when card was non-3DS")
	}

	_, err = client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             amount,
		PaymentMethodNonce: nonce.Nonce,
		Options: &TxOpts{
			ThreeDSecure: &TxThreeDSecureOptsRequest{Required: true},
		},
	})
	if err == nil {
		t.Fatal("Did not receive error when creating transaction requiring 3DS with non-3DS nonce")
	}
	if err.Error() != "Gateway Rejected: three_d_secure" {
		t.Fatal(err)
	}
	if err.(*APIError).Transaction.ThreeDSecureInfo != nil {
		t.Fatalf("Tx 3DS Info present when nonce for transaction was non-3DS")
	}
}

func randomAmount() *Decimal {
	return NewDecimal(1+rand.Int63n(9999), 2)
}

func TestTransactionCreateSubmitForSettlementAndVoid(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(2000, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	// Submit for settlement
	ten := NewDecimal(1000, 2)
	tx2, err := client.SubmitForSettlement(context.Background(), tx.Id, ten)

	if err != nil {
		t.Fatal(err)
	}
	if x := tx2.Status; x != StatusSubmittedForSettlement {
		t.Fatal(x)
	}
	if amount := tx2.Amount; amount.Cmp(ten) != 0 {
		t.Fatalf("transaction settlement amount (%s) did not equal amount requested (%s)", amount, ten)
	}

	// Void
	tx3, err := client.Void(context.Background(), tx2.Id)

	if err != nil {
		t.Fatal(err)
	}
	if x := tx3.Status; x != StatusVoided {
		t.Fatal(x)
	}
}

func TestTransactionSearchIDs(t *testing.T) {
	t.Parallel()

	createTx := func(amount *Decimal, customerName string) (*Tx, error) {
		return client.Pay(context.Background(), &TxRequest{
			Type:   "sale",
			Amount: amount,
			Customer: &CustomerRequest{
				FirstName: customerName,
			},
			CreditCard: &CreditCard{
				Number:         testCardVisa,
				ExpirationDate: "05/14",
			},
		})
	}

	unique := RandomString()

	name0 := "Erik-" + unique
	tx1, err := createTx(randomAmount(), name0)
	if err != nil {
		t.Fatal(err)
	}

	name1 := "Lionel-" + unique
	_, err = createTx(randomAmount(), name1)
	if err != nil {
		t.Fatal(err)
	}

	query := new(Search)
	f := query.AddTextField("customer-first-name")
	f.Is = name0

	result, err := client.SearchTxs(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.IDs) != 1 {
		t.Fatal(result.IDs)
	}

	if tx1.Id != result.IDs[0] {
		t.Fatal(result)
	}
}

func TestTransactionSearchPage(t *testing.T) {
	t.Parallel()

	const transactionCount = 51
	transactionIDs := map[string]bool{}
	prefix := "PaginationTest-" + RandomString()
	for i := 0; i < transactionCount; i++ {
		unique := RandomString()
		tx, err := client.Pay(context.Background(), &TxRequest{
			Type:   "sale",
			Amount: randomAmount(),
			Customer: &CustomerRequest{
				FirstName: prefix + unique,
			},
			CreditCard: &CreditCard{
				Number:         testCardVisa,
				ExpirationDate: "05/14",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		transactionIDs[tx.Id] = true
	}

	query := new(Search)
	query.AddTextField("customer-first-name").StartsWith = prefix

	results, err := client.SearchTxs(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.IDs) != transactionCount {
		t.Fatalf("results.IDs = %v, want %v", len(results.IDs), transactionCount)
	}

	for page := 1; page <= results.PageCount; page++ {
		results.Page = page
		resultz, err := client.SearchTx(context.Background(), query, results)
		if err != nil {
			t.Fatal(err)
		}
		for _, tx := range resultz.Transactions {
			if firstName := tx.Customer.FirstName; !strings.HasPrefix(firstName, prefix) {
				t.Fatalf("tx.Customer.FirstName = %q, want prefix of %q", firstName, prefix)
			}
			if transactionIDs[tx.Id] {
				delete(transactionIDs, tx.Id)
			} else {
				t.Fatalf("tx.Id = %q, not expected", tx.Id)
			}
		}
	}

	if len(transactionIDs) > 0 {
		t.Fatalf("transactions not returned = %v", transactionIDs)
	}
	results.Page = 0
	_, err = client.SearchTx(context.Background(), query, results)
	if err == nil || !strings.Contains(err.Error(), "page 0 out of bounds") {
		t.Errorf("requesting page 0 should result in out of bounds error, but got %#v", err)
	}
	results.Page = results.PageCount + 1
	_, err = client.SearchTx(context.Background(), query, results)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("page %d out of bounds", results.PageCount+1)) {
		t.Errorf("requesting page %d should result in out of bounds error, but got %v", results.PageCount+1, err)
	}
}

func TestTransactionSearch(t *testing.T) {
	t.Parallel()

	createTx := func(amount *Decimal, customerName string) error {
		_, err := client.Pay(context.Background(), &TxRequest{
			Type:   "sale",
			Amount: amount,
			Customer: &CustomerRequest{
				FirstName: customerName,
			},
			CreditCard: &CreditCard{
				Number:         testCardVisa,
				ExpirationDate: "05/14",
			},
		})
		return err
	}

	unique := RandomString()

	name0 := "Erik-" + unique
	if err := createTx(randomAmount(), name0); err != nil {
		t.Fatal(err)
	}

	name1 := "Lionel-" + unique
	if err := createTx(randomAmount(), name1); err != nil {
		t.Fatal(err)
	}

	query := new(Search)
	f := query.AddTextField("customer-first-name")
	f.Is = name0

	searchResult, err := client.SearchTxs(context.Background(), query)
	if err != nil {
		t.Fatalf("Error : %v", err)
	}

	pageSize := searchResult.PageSize
	ids := searchResult.IDs

	endOffset := pageSize
	if endOffset > len(ids) {
		endOffset = len(ids)
	}

	firstPageQuery := query.ShallowCopy()
	firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
	firstPageTransactions, err := client.FetchTx(context.Background(), firstPageQuery)

	result := &TransactionSearchResult{
		TotalItems:        len(ids),
		TotalIDs:          ids,
		CurrentPageNumber: 1,
		PageSize:          pageSize,
		Transactions:      firstPageTransactions,
	}

	if result.TotalItems != 1 {
		t.Fatal(result.Transactions)
	}

	tx := result.Transactions[0]
	if x := tx.Customer.FirstName; x != name0 {
		t.Fatal(x)
	}
}

func TestTransactionSearchTime(t *testing.T) {

	createTx := func(amount *Decimal, customerName string) error {
		_, err := client.Pay(context.Background(), &TxRequest{
			Type:   "sale",
			Amount: amount,
			Customer: &CustomerRequest{
				FirstName: customerName,
			},
			CreditCard: &CreditCard{
				Number:         testCardVisa,
				ExpirationDate: "05/14",
			},
		})
		return err
	}

	unique := RandomString()

	name0 := "Erik-" + unique
	if err := createTx(randomAmount(), name0); err != nil {
		t.Fatal(err)
	}

	name1 := "Lionel-" + unique
	if err := createTx(randomAmount(), name1); err != nil {
		t.Fatal(err)
	}

	{ // test: txn is returned if querying for created at before now
		query := new(Search)
		f1 := query.AddTextField("customer-first-name")
		f1.Is = name0
		f2 := query.AddTimeField("created-at")
		f2.Max = time.Now()

		searchResult, err := client.SearchTxs(context.Background(), query)
		if err != nil {
			t.Fatalf("Error : %v", err)
		}

		pageSize := searchResult.PageSize
		ids := searchResult.IDs

		endOffset := pageSize
		if endOffset > len(ids) {
			endOffset = len(ids)
		}

		firstPageQuery := query.ShallowCopy()
		firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
		firstPageTransactions, err := client.FetchTx(context.Background(), firstPageQuery)

		result := &TransactionSearchResult{
			TotalItems:        len(ids),
			TotalIDs:          ids,
			CurrentPageNumber: 1,
			PageSize:          pageSize,
			Transactions:      firstPageTransactions,
		}

		if result.TotalItems != 1 {
			t.Fatal(result.Transactions)
		}

		tx := result.Transactions[0]
		if x := tx.Customer.FirstName; x != name0 {
			t.Fatal(x)
		}
	}

	{ // test: txn is not returned if querying for created at before 1 hour ago
		query := new(Search)
		f1 := query.AddTextField("customer-first-name")
		f1.Is = name0
		f2 := query.AddTimeField("created-at")
		f2.Max = time.Now().Add(-time.Hour)

		searchResult, err := client.SearchTxs(context.Background(), query)
		if err != nil {
			t.Fatalf("Error : %v", err)
		}

		pageSize := searchResult.PageSize
		ids := searchResult.IDs

		endOffset := pageSize
		if endOffset > len(ids) {
			endOffset = len(ids)
		}

		firstPageQuery := query.ShallowCopy()
		firstPageQuery.AddMultiField("ids").Items = ids[:endOffset]
		firstPageTransactions, err := client.FetchTx(context.Background(), firstPageQuery)

		result := &TransactionSearchResult{
			TotalItems:        len(ids),
			TotalIDs:          ids,
			CurrentPageNumber: 1,
			PageSize:          pageSize,
			Transactions:      firstPageTransactions,
		}

		if result.TotalItems != 0 {
			t.Fatal(result.Transactions)
		}
	}
}

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestTransactionCreateWhenGatewayRejected(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(201000, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}
	if err.Error() != "Card Issuer Declined CVV" {
		t.Fatal(err)
	}
	if err.(*APIError).Transaction.ProcessorResponseCode != 2010 {
		t.Fatalf("expected err.Tx.ResponseCode to be 2010, but got %d", err.(*APIError).Transaction.ProcessorResponseCode)
	}
	if err.(*APIError).Transaction.ProcessorResponseType != ResponseTypeHardDeclined {
		t.Fatalf("expected err.Tx.ResponseType to be %s, but got %s", ResponseTypeHardDeclined, err.(*APIError).Transaction.ProcessorResponseType)
	}

	if err.(*APIError).Transaction.AdditionalProcessorResponse != "2010 : Card Issuer Declined CVV" {
		t.Fatalf("expected err.Tx.ResponseCode to be `2010 : Card Issuer Declined CVV`, but got %s", err.(*APIError).Transaction.AdditionalProcessorResponse)
	}
}

func TestTransactionCreateWhenGatewayRejectedFraud(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(201000, 2),
		PaymentMethodNonce: FakeNonceGatewayRejectedFraud,
	})
	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}

	if err.Error() != "Gateway Rejected: fraud" {
		t.Fatal(err)
	}

	txn := err.(*APIError).Transaction
	if txn.Status != StatusGatewayRejected {
		t.Fatalf("Got status %q, want %q", txn.Status, StatusGatewayRejected)
	}

	if txn.GatewayRejectionReason != FraudReason {
		t.Fatalf("Got gateway rejection reason %q, wanted %q", txn.GatewayRejectionReason, FraudReason)
	}

	if txn.ProcessorResponseCode != 0 {
		t.Fatalf("Got processor response code %q, want %q", txn.ProcessorResponseCode, 0)
	}
}

func TestTransactionCreatedWhenCVVDoesNotMatch(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "200", // Should cause CVV does not match response
		},
	})

	if err.Error() != "Gateway Rejected: cvv" {
		t.Fatal(err)
	}

	txn := err.(*APIError).Transaction

	if txn.Status != StatusGatewayRejected {
		t.Fatalf("Got status %q, want %q", txn.Status, StatusGatewayRejected)
	}

	if txn.GatewayRejectionReason != CVVReason {
		t.Fatalf("Got gateway rejection reason %q, wanted %q", txn.GatewayRejectionReason, CVVReason)
	}

	if txn.CVVResponseCode != CVVResponseCodeDoesNotMatch {
		t.Fatalf("Got CVV Response Code %q, wanted %q", txn.CVVResponseCode, CVVResponseCodeDoesNotMatch)
	}
}

func TestTransactionCreatedWhenAVSBankDoesNotSupport(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		MerchantAccountId: avsAndCVVTestMerchantAccountId,
		Type:              "sale",
		Amount:            randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "100",
		},
		BillingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "30001", // Should cause AVS bank does not support error response.
		},
	})

	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}

	if err.Error() != "Gateway Rejected: avs" {
		t.Fatal(err)
	}

	txn := err.(*APIError).Transaction

	if txn.Status != StatusGatewayRejected {
		t.Fatalf("Got status %q, want %q", txn.Status, StatusGatewayRejected)
	}

	if txn.GatewayRejectionReason != AVSReason {
		t.Fatalf("Got gateway rejection reason %q, wanted %q", txn.GatewayRejectionReason, AVSReason)
	}

	if txn.AVSErrorResponseCode != AVSResponseCodeNotSupported {
		t.Fatalf("Got AVS Error Response Code %q, wanted %q", txn.AVSErrorResponseCode, AVSResponseCodeNotSupported)
	}
}

func TestTransactionCreatedWhenAVSPostalDoesNotMatch(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		MerchantAccountId: avsAndCVVTestMerchantAccountId,
		Type:              "sale",
		Amount:            randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "100",
		},
		BillingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "20000", // Should cause AVS postal code does not match response.
		},
	})

	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}

	if err.Error() != "Gateway Rejected: avs" {
		t.Fatal(err)
	}

	txn := err.(*APIError).Transaction

	if txn.Status != StatusGatewayRejected {
		t.Fatalf("Got status %q, want %q", txn.Status, StatusGatewayRejected)
	}

	if txn.GatewayRejectionReason != AVSReason {
		t.Fatalf("Got gateway rejection reason %q, wanted %q", txn.GatewayRejectionReason, AVSReason)
	}

	if txn.AVSPostalCodeResponseCode != AVSResponseCodeDoesNotMatch {
		t.Fatalf("Got AVS postal response code %q, wanted %q", txn.AVSPostalCodeResponseCode, AVSResponseCodeDoesNotMatch)
	}
}

func TestTransactionCreatedWhenAVStreetAddressDoesNotMatch(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		MerchantAccountId: avsAndCVVTestMerchantAccountId,
		Type:              "sale",
		Amount:            randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "100",
		},
		BillingAddress: &Address{
			StreetAddress: "201 E Main St", // Should cause AVS street address not verified response.
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "60637",
		},
	})

	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}

	if err.Error() != "Gateway Rejected: avs" {
		t.Fatal(err)
	}

	txn := err.(*APIError).Transaction

	if txn.Status != StatusGatewayRejected {
		t.Fatalf("Got status %q, want %q", txn.Status, StatusGatewayRejected)
	}

	if txn.GatewayRejectionReason != AVSReason {
		t.Fatalf("Got gateway rejection reason %q, wanted %q", txn.GatewayRejectionReason, AVSReason)
	}

	if txn.AVSStreetAddressResponseCode != AVSResponseCodeNotVerified {
		t.Fatalf("Got AVS street address response code %q, wanted %q", txn.AVSStreetAddressResponseCode, AVSResponseCodeNotVerified)
	}
}

func TestFindTransaction(t *testing.T) {
	t.Parallel()

	createdTransaction, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardMastercard,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	foundTransaction, err := client.FindTransaction(context.Background(), createdTransaction.Id)
	if err != nil {
		t.Fatal(err)
	}

	if createdTransaction.Id != foundTransaction.Id {
		t.Fatal("transaction ids do not match")
	}
}

func TestFindNonExistantTransaction(t *testing.T) {
	t.Parallel()

	_, err := client.FindTransaction(context.Background(), "bad_transaction_id")
	if err == nil {
		t.Fatal("Did not receive error when finding an invalid tx ID")
	}
	if err.Error() != "Not Found (404)" {
		t.Fatal(err)
	}
	if apiErr, ok := err.(IAPIError); !(ok && apiErr.StatusCode() == http.StatusNotFound) {
		t.Fatal(err)
	}
}

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestTransactionDescriptorFields(t *testing.T) {
	t.Parallel()

	tx := &TxRequest{
		Type:               "sale",
		Amount:             randomAmount(),
		PaymentMethodNonce: FakeNonceTransactable,
		Options: &TxOpts{
			SubmitForSettlement: true,
		},
		Descriptor: &Descriptor{
			Name:  "Company Name*Product 1",
			Phone: "0000000000",
			URL:   "example.com",
		},
	}

	tx2, err := client.Pay(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if tx2.Type != tx.Type {
		t.Fatalf("expected Type to be equal, but %s was not %s", tx2.Type, tx.Type)
	}
	if tx2.Amount.Cmp(tx.Amount) != 0 {
		t.Fatalf("expected Amount to be equal, but %s was not %s", tx2.Amount, tx.Amount)
	}
	if tx2.Status != StatusSubmittedForSettlement {
		t.Fatalf("expected tx2.Status to be %s, but got %s", StatusSubmittedForSettlement, tx2.Status)
	}
	if tx2.Descriptor.Name != "Company Name*Product 1" {
		t.Fatalf("expected tx2.Descriptor.Name to be Company Name*Product 1, but got %s", tx2.Descriptor.Name)
	}
	if tx2.Descriptor.Phone != "0000000000" {
		t.Fatalf("expected tx2.Descriptor.Phone to be 0000000000, but got %s", tx2.Descriptor.Phone)
	}
	if tx2.Descriptor.URL != "example.com" {
		t.Fatalf("expected tx2.Descriptor.URL to be example.com, but got %s", tx2.Descriptor.URL)
	}
}

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestTransactionPaypalFields(t *testing.T) {
	t.Parallel()

	const (
		PayeeEmail  = "payee@payal.com"
		Description = "One tasty sandwich"
		CustomField = "foo"
	)
	subData := make(map[string]string)
	subData["faz"] = "bar"

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{})
	if err != nil {
		t.Fatal(err)
	}

	nonce := FakeNoncePayPalFuturePayment

	paymentMethod, err := client.CreatePayMethod(context.Background(), &PaymentMethodRequest{
		CustomerId:         customer.Id,
		PaymentMethodNonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	tx := &TxRequest{
		Type:               "sale",
		Amount:             randomAmount(),
		PaymentMethodToken: paymentMethod.Token,
		OrderId:            "123456ABC",
		Options: &TxOpts{
			SubmitForSettlement: true,
			TransactionOptionsPaypalRequest: &TxPaypalOptsRequest{
				PayeeEmail:        PayeeEmail,
				Description:       Description,
				CustomField:       CustomField,
				SupplementaryData: subData,
			},
		},
	}
	tx2, err := client.Pay(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	if tx2.Type != tx.Type {
		t.Fatalf("expected Type to be equal, but %s was not %s", tx2.Type, tx.Type)
	}
	if tx2.Amount.Cmp(tx.Amount) != 0 {
		t.Fatalf("expected Amount to be equal, but %s was not %s", tx2.Amount, tx.Amount)
	}
	if tx2.Status != StatusSettling {
		t.Fatalf("expected tx2.Status to be %s, but got %s", StatusSettling, tx2.Status)
	}
	if tx2.PayPalDetails.PayeeEmail != PayeeEmail {
		t.Fatalf("expected tx2.PaypalDetails.PayeeEmail to be %s, but got %s", PayeeEmail, tx2.PayPalDetails.PayeeEmail)
	}
	if tx2.PayPalDetails.Description != Description {
		t.Fatalf("expected tx2.PaypalDetails.Description to be %s, but got %s", Description, tx2.PayPalDetails.Description)
	}
	if tx2.PayPalDetails.CustomField != CustomField {
		t.Fatalf("expected tx2.PayPalDetail.CustomField to be %s, but got %s", CustomField, tx2.PayPalDetails.CustomField)
	}
}

func TestTransactionRiskDataFields(t *testing.T) {
	t.Parallel()

	tx := &TxRequest{
		Type:               "sale",
		Amount:             randomAmount(),
		PaymentMethodNonce: FakeNonceTransactable,
		RiskData: &RiskDataRequest{
			CustomerBrowser: "Mozilla/5.0 (X11; U; Linux x86_64; en-US) AppleWebKit/540.0 (KHTML,like Gecko) Chrome/9.1.0.0 Safari/540.0",
			CustomerIP:      "127.0.0.1",
		},
	}

	tx2, err := client.Pay(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if tx2.Type != tx.Type {
		t.Fatalf("expected Type to be equal, but %s was not %s", tx2.Type, tx.Type)
	}
	if tx2.Amount.Cmp(tx.Amount) != 0 {
		t.Fatalf("expected Amount to be equal, but %s was not %s", tx2.Amount, tx.Amount)
	}
}

func TestTransactionSkipAdvancedFraudChecks(t *testing.T) {
	t.Parallel()

	tx := &TxRequest{
		Type:               "sale",
		Amount:             randomAmount(),
		PaymentMethodNonce: FakeNonceTransactable,
		RiskData: &RiskDataRequest{
			CustomerBrowser: "Mozilla/5.0 (X11; U; Linux x86_64; en-US) AppleWebKit/540.0 (KHTML,like Gecko) Chrome/9.1.0.0 Safari/540.0",
			CustomerIP:      "127.0.0.1",
		},
		Options: &TxOpts{
			SkipAdvancedFraudChecking: true,
		},
	}

	tx2, err := client.Pay(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	if tx2.RiskData != nil {
		t.Fatal("expected tx2.RiskData to be empty")
	}
}

func TestAllTransactionFields(t *testing.T) {
	t.Parallel()

	amount := randomAmount()
	taxAmount := NewDecimal(amount.Unscaled/10, amount.Scale)

	tx := &TxRequest{
		Type:    "sale",
		Amount:  amount,
		OrderId: "my_custom_order",
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
			CVV:            "100",
		},
		TransactionSource: MOTO,
		Customer: &CustomerRequest{
			FirstName: "Lionel",
		},
		BillingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "60637",
		},
		ShippingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "60637",
		},
		TaxAmount:           taxAmount,
		DeviceData:          `{"device_session_id": "dsi_1234", "fraud_merchant_id": "fmi_1234", "correlation_id": "ci_1234"}`,
		Channel:             "ChannelA",
		PurchaseOrderNumber: "PONUMBER",
		Options: &TxOpts{
			SubmitForSettlement:              true,
			StoreInVault:                     true,
			StoreInVaultOnSuccess:            true,
			AddBillingAddressToPaymentMethod: true,
			StoreShippingAddressInVault:      true,
		},
	}

	tx2, err := client.Pay(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if tx2.Type != tx.Type {
		t.Fatalf("expected Type to be equal, but %s was not %s", tx2.Type, tx.Type)
	}
	if tx2.CurrencyISOCode != "USD" {
		t.Fatalf("expected CurrencyISOCode to be %s but was %s", "USD", tx2.CurrencyISOCode)
	}
	if tx2.Amount.Cmp(tx.Amount) != 0 {
		t.Fatalf("expected Amount to be equal, but %s was not %s", tx2.Amount, tx.Amount)
	}
	if tx2.OrderId != tx.OrderId {
		t.Fatalf("expected OrderId to be equal, but %s was not %s", tx2.OrderId, tx.OrderId)
	}
	if tx2.Customer.FirstName != tx.Customer.FirstName {
		t.Fatalf("expected Customer.FirstName to be equal, but %s was not %s", tx2.Customer.FirstName, tx.Customer.FirstName)
	}
	if tx2.BillingAddress.StreetAddress != tx.BillingAddress.StreetAddress {
		t.Fatalf("expected BillingAddress.StreetAddress to be equal, but %s was not %s", tx2.BillingAddress.StreetAddress, tx.BillingAddress.StreetAddress)
	}
	if tx2.BillingAddress.Locality != tx.BillingAddress.Locality {
		t.Fatalf("expected BillingAddress.Locality to be equal, but %s was not %s", tx2.BillingAddress.Locality, tx.BillingAddress.Locality)
	}
	if tx2.BillingAddress.Region != tx.BillingAddress.Region {
		t.Fatalf("expected BillingAddress.Region to be equal, but %s was not %s", tx2.BillingAddress.Region, tx.BillingAddress.Region)
	}
	if tx2.BillingAddress.PostalCode != tx.BillingAddress.PostalCode {
		t.Fatalf("expected BillingAddress.PostalCode to be equal, but %s was not %s", tx2.BillingAddress.PostalCode, tx.BillingAddress.PostalCode)
	}
	if tx2.ShippingAddress.StreetAddress != tx.ShippingAddress.StreetAddress {
		t.Fatalf("expected ShippingAddress.StreetAddress to be equal, but %s was not %s", tx2.ShippingAddress.StreetAddress, tx.ShippingAddress.StreetAddress)
	}
	if tx2.ShippingAddress.Locality != tx.ShippingAddress.Locality {
		t.Fatalf("expected ShippingAddress.Locality to be equal, but %s was not %s", tx2.ShippingAddress.Locality, tx.ShippingAddress.Locality)
	}
	if tx2.ShippingAddress.Region != tx.ShippingAddress.Region {
		t.Fatalf("expected ShippingAddress.Region to be equal, but %s was not %s", tx2.ShippingAddress.Region, tx.ShippingAddress.Region)
	}
	if tx2.ShippingAddress.PostalCode != tx.ShippingAddress.PostalCode {
		t.Fatalf("expected ShippingAddress.PostalCode to be equal, but %s was not %s", tx2.ShippingAddress.PostalCode, tx.ShippingAddress.PostalCode)
	}
	if tx2.TaxAmount == nil {
		t.Fatalf("expected TaxAmount to be set, but was nil")
	}
	if tx2.TaxAmount.Cmp(tx.TaxAmount) != 0 {
		t.Fatalf("expected TaxAmount to be equal, but %s was not %s", tx2.TaxAmount, tx.TaxAmount)
	}
	if tx2.TaxExempt != tx.TaxExempt {
		t.Fatalf("expected TaxExempt to be equal, but %t was not %t", tx2.TaxExempt, tx.TaxExempt)
	}
	if tx2.CreditCard.Token == "" {
		t.Fatalf("expected CreditCard.Token to be equal, but %s was not %s", tx2.CreditCard.Token, tx.CreditCard.Token)
	}
	if tx2.Customer.Id == "" {
		t.Fatalf("expected Customer.Id to be equal, but %s was not %s", tx2.Customer.Id, tx.Customer.ID)
	}
	if tx2.Status != StatusSubmittedForSettlement {
		t.Fatalf("expected tx2.Status to be %s, but got %s", StatusSubmittedForSettlement, tx2.Status)
	}
	if tx2.PaymentInstrumentType != CreditCardType {
		t.Fatalf("expected tx2.PaymentType to be %s, but got %s", CreditCardType, tx2.PaymentInstrumentType)
	}
	if tx2.AdditionalProcessorResponse != "" {
		t.Fatalf("expected tx2.AdditionalProcessorResponse to be empty, but got %s", tx2.AdditionalProcessorResponse)
	}
	if tx2.ProcessorResponseType != ResponseTypeApproved {
		t.Fatalf("expected tx2.ResponseType to be %s, but got %s", ResponseTypeApproved, tx2.ProcessorResponseType)
	}

	if tx2.RiskData != nil {
		switch tx2.RiskData.Decision {
		case "Not Evaluated":
			if tx2.RiskData.ID != "" {
				t.Fatalf("expected tx2.RiskData.ID to be empty when Decision is Not Evaluated, but got %q", tx2.RiskData.ID)
			}
		case "Approve":
			if tx2.RiskData.ID == "" {
				t.Fatalf("expected tx2.RiskData.ID to be non-empty when Decision is Approved, but got %q", tx2.RiskData.ID)
			}
		default:
			t.Fatalf("expected tx2.RiskData.Decision to be Not Evaluated or Approved, but got %s", tx2.RiskData.Decision)
		}
	}
	if tx2.Channel != "ChannelA" {
		t.Fatalf("expected tx2.Channel to be ChannelA, but got %s", tx2.Channel)
	}
	if tx2.PurchaseOrderNumber != tx.PurchaseOrderNumber {
		t.Fatalf("expected PurchaseOrderNumber to be %s, but got %s", tx.PurchaseOrderNumber, tx2.PurchaseOrderNumber)
	}
	if tx2.SubscriptionDetails != nil {
		t.Fatalf("expected Subscription to be not nil, but got %#v", tx2.SubscriptionDetails)
	}
	if tx2.AuthorizationExpiresAt == nil {
		t.Fatalf("expected AuthorizationExpiresAt to be not nil, but got %#v", tx2.AuthorizationExpiresAt)
	} else if tx2.AuthorizationExpiresAt.Before(time.Now()) || tx2.AuthorizationExpiresAt.After(time.Now().AddDate(0, 0, 60)) {
		t.Fatalf("expected AuthorizationExpiresAt to be between the current time and 60 days from now, but got %s", tx2.AuthorizationExpiresAt.Format(time.RFC3339))
	}
}

// This test will only pass on Travis. See TESTING.md for more details.
func TestTransactionDisbursementDetails(t *testing.T) {
	t.Parallel()

	txn, err := client.FindTransaction(context.Background(), "dskdmb")
	if err != nil {
		t.Fatal(err)
	}

	if txn.DisbursementDetails.DisbursementDate != "2013-06-27" {
		t.Fatalf("expected disbursement date to be %s, was %s", "2013-06-27", txn.DisbursementDetails.DisbursementDate)
	}
	if txn.DisbursementDetails.SettlementAmount.Cmp(NewDecimal(10000, 2)) != 0 {
		t.Fatalf("expected settlement amount to be %s, was %s", NewDecimal(10000, 2), txn.DisbursementDetails.SettlementAmount)
	}
	if txn.DisbursementDetails.SettlementCurrencyIsoCode != "USD" {
		t.Fatalf("expected settlement currency code to be %s, was %s", "USD", txn.DisbursementDetails.SettlementCurrencyIsoCode)
	}
	if txn.DisbursementDetails.SettlementCurrencyExchangeRate.Cmp(NewDecimal(100, 2)) != 0 {
		t.Fatalf("expected settlement currency exchange rate to be %s, was %s", NewDecimal(100, 2), txn.DisbursementDetails.SettlementCurrencyExchangeRate)
	}
	if txn.DisbursementDetails.FundsHeld {
		t.Error("funds held doesn't match")
	}
	if !txn.DisbursementDetails.Success {
		t.Error("success doesn't match")
	}
}

func TestTransactionCreateFromPaymentMethodCode(t *testing.T) {
	t.Parallel()

	customer, err := client.CreateCustomer(context.Background(), &CustomerRequest{
		CreditCard: &CreditCard{
			Number:         testCardDiscover,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if customer.CreditCards.CreditCard[0].Token == "" {
		t.Fatal("invalid token")
	}

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		CustomerID:         customer.Id,
		Amount:             randomAmount(),
		PaymentMethodToken: customer.CreditCards.CreditCard[0].Token,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("invalid tx id")
	}
}

func TestTrxPaymentMethodNonce(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             randomAmount(),
		PaymentMethodNonce: "fake-apple-pay-mastercard-nonce",
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransactionCreateSettleAndFullRefund(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(20000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	// Refund
	refundTxn, err := client.Refund(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}
	if x := refundTxn.Status; x != StatusSubmittedForSettlement {
		t.Fatal(x)
	}

	refundTxn, err = client.SandboxSettle(context.Background(), refundTxn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if refundTxn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	if *refundTxn.RefundedTransactionId != txn.Id {
		t.Fatal(*refundTxn.RefundedTransactionId)
	}

	// Check that the refund shows up in the original transaction
	txn, err = client.FindTransaction(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.RefundIds != nil && (*txn.RefundIds)[0] != refundTxn.Id {
		t.Fatal(*txn.RefundIds)
	}

	// Second refund should fail
	refundTxn, err = client.Refund(context.Background(), txn.Id)

	if err.Error() != "Tx has already been fully refunded." {
		t.Fatal(err)
	}
}

func TestTransactionCreateSettleAndFullRefundWithRequest(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(20000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	// Refund
	refundTxn, err := client.RefundWithRequest(context.Background(), txn.Id, &RefundRequest{
		OrderID: "fully-refunded-tx",
	})

	if err != nil {
		t.Fatal(err)
	}
	if x := refundTxn.Status; x != StatusSubmittedForSettlement {
		t.Fatal(x)
	}

	refundTxn, err = client.SandboxSettle(context.Background(), refundTxn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if refundTxn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	if *refundTxn.RefundedTransactionId != txn.Id {
		t.Fatal(*refundTxn.RefundedTransactionId)
	}

	if refundTxn.OrderId != "fully-refunded-tx" {
		t.Fatal(refundTxn.OrderId)
	}

	// Check that the refund shows up in the original transaction
	txn, err = client.FindTransaction(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.RefundIds != nil && (*txn.RefundIds)[0] != refundTxn.Id {
		t.Fatal(*txn.RefundIds)
	}

	// Second refund should fail
	refundTxn, err = client.RefundWithRequest(context.Background(), txn.Id, &RefundRequest{
		OrderID: "fully-refunded-tx",
	})

	if err.Error() != "Tx has already been fully refunded." {
		t.Fatal(err)
	}
}

func TestTransactionCreateSettleAndPartialRefund(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(10000, 2)
	refundAmt1 := NewDecimal(5000, 2)
	refundAmt2 := NewDecimal(5001, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	// Refund
	refundTxn, err := client.Refund(context.Background(), txn.Id, refundAmt1)

	if err != nil {
		t.Fatal(err)
	}
	if x := refundTxn.Status; x != StatusSubmittedForSettlement {
		t.Fatal(x)
	}

	refundTxn, err = client.SandboxSettle(context.Background(), refundTxn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if refundTxn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}

	// Refund amount too large
	refundTxn, err = client.Refund(context.Background(), txn.Id, refundAmt2)

	if err.Error() != "Refund amount is too large." {
		t.Fatal(err)
	}
}

func TestTransactionCreateWithCustomFields(t *testing.T) {
	t.Parallel()

	customFields := map[string]string{
		"custom_field_1": "custom value",
	}

	amount := NewDecimal(10000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             amount,
		PaymentMethodNonce: FakeNonceTransactable,
		CustomFields:       customFields,
	})
	if err != nil {
		t.Fatal(err)
	}

	if x := map[string]string(txn.CustomFields); !reflect.DeepEqual(x, customFields) {
		t.Fatalf("Returned custom fields doesn't match input, got %q, want %q", x, customFields)
	}

	txn, err = client.FindTransaction(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if x := map[string]string(txn.CustomFields); !reflect.DeepEqual(x, customFields) {
		t.Fatalf("Returned custom fields doesn't match input, got %q, want %q", x, customFields)
	}
}

func TestTransactionTaxExempt(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(10000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             amount,
		TaxExempt:          true,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.FindTransaction(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !txn.TaxExempt {
		t.Fatalf("Tx did not return tax exempt")
	}
	if txn.TaxAmount != nil {
		t.Fatalf("Tx TaxAmount got %v, want nil", txn.TaxAmount)
	}
}

func TestTransactionTaxFieldsNotProvided(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(10000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             amount,
		PaymentMethodNonce: FakeNonceTransactable,
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.FindTransaction(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.TaxExempt {
		t.Fatalf("Tx returned tax exempt, expected not to")
	}
	if txn.TaxAmount != nil {
		t.Fatalf("Tx tax amount got %v, want nil", *txn.TaxAmount)
	}
}

func TestEscrowHoldOnCreate(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6200, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
		Options: &TxOpts{
			HoldInEscrow: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.EscrowStatus != EscrowHoldPending {
		t.Fatalf("Tx EscrowStatus got %s, want %s", txn.EscrowStatus, EscrowHoldPending)
	}
}

func TestEscrowHoldOnCreateOnMasterMerchant(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6301, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		Options: &TxOpts{
			HoldInEscrow: true,
		},
	})
	if err == nil {
		t.Fatal("Tx Sale got no error, want error")
	}
	errors := err.(*APIError).For("Transaction").On("Base")
	if len(errors) != 1 {
		t.Fatalf("Tx Sale got %d errors, want 1 error", len(errors))
	}
	if g, w := errors[0].Code, "91560"; g != w {
		t.Errorf("Tx Sale got error code %s, want %s", g, w)
	}
	if g, w := errors[0].Message, "Transaction could not be held in escrow."; g != w {
		t.Errorf("Tx Sale got error message %s, want %s", g, w)
	}
}

func TestEscrowHoldAfterSale(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6300, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	txn, err = client.HoldInEscrow(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	if txn.EscrowStatus != EscrowHoldPending {
		t.Fatalf("Tx EscrowStatus got %s, want %s", txn.EscrowStatus, EscrowHoldPending)
	}
}

func TestEscrowHoldAfterSaleOnMasterMerchant(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6301, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.HoldInEscrow(context.Background(), txn.Id)
	if err == nil {
		t.Fatal("Tx HoldInEscrow got no error, want error")
	}
	errors := err.(*APIError).For("Transaction").On("Base")
	if len(errors) != 1 {
		t.Fatalf("Tx HoldInEscrow got %d errors, want 1 error", len(errors))
	}
	if g, w := errors[0].Code, "91560"; g != w {
		t.Errorf("Tx HoldInEscrow got error code %s, want %s", g, w)
	}
	if g, w := errors[0].Message, "Transaction could not be held in escrow."; g != w {
		t.Errorf("Tx HoldInEscrow got error message %s, want %s", g, w)
	}
}

func TestEscrowRelease(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6400, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
		Options: &TxOpts{
			SubmitForSettlement: true,
			HoldInEscrow:        true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	txn, err = client.ReleaseFromEscrow(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	if txn.EscrowStatus != EscrowReleasePending {
		t.Fatalf("Tx EscrowStatus got %s, want %s", txn.EscrowStatus, EscrowReleasePending)
	}
}

func TestEscrowReleaseNotEscrowed(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6401, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ReleaseFromEscrow(context.Background(), txn.Id)
	if err == nil {
		t.Fatal("Tx ReleaseFromEscrow got no error, want error")
	}
	errors := err.(*APIError).For("Transaction").On("Base")
	if len(errors) != 1 {
		t.Fatalf("Tx ReleaseFromEscrow got %d errors, want 1 error", len(errors))
	}
	if g, w := errors[0].Code, "91561"; g != w {
		t.Errorf("Tx ReleaseFromEscrow got error code %s, want %s", g, w)
	}
	if g, w := errors[0].Message, "Cannot release a transaction that is not escrowed."; g != w {
		t.Errorf("Tx ReleaseFromEscrow got error message %s, want %s", g, w)
	}
}

func TestEscrowCancelRelease(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6500, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
		Options: &TxOpts{
			SubmitForSettlement: true,
			HoldInEscrow:        true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	txn, err = client.ReleaseFromEscrow(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	if txn.EscrowStatus != EscrowReleasePending {
		t.Fatalf("Tx EscrowStatus got %s, want %s", txn.EscrowStatus, EscrowReleasePending)
	}
	txn, err = client.CancelRelease(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}
	if txn.EscrowStatus != EscrowHeld {
		t.Fatalf("Tx EscrowStatus got %s, want %s", txn.EscrowStatus, EscrowHeld)
	}
}

func TestEscrowCancelReleaseNotPending(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(6501, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
		MerchantAccountId: testSubMerchantAccount(),
		ServiceFeeAmount:  NewDecimal(1000, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CancelRelease(context.Background(), txn.Id)
	if err == nil {
		t.Fatal("Tx Cancel Release got no error, want error")
	}
	errors := err.(*APIError).For("Transaction").On("Base")
	if len(errors) != 1 {
		t.Fatalf("Tx Cancel Release got %d errors, want 1 error", len(errors))
	}
	if g, w := errors[0].Code, "91562"; g != w {
		t.Errorf("Tx Cancel Release got error code %s, want %s", g, w)
	}
	if g, w := errors[0].Message, "Release can only be cancelled if the transaction is submitted for release."; g != w {
		t.Errorf("Tx Cancel Release got error message %s, want %s", g, w)
	}
}

func TestTransactionStoreInVault(t *testing.T) {
	t.Parallel()

	type args struct {
		request *TxRequest
	}
	tests := []struct {
		name      string
		args      args
		wantToken bool
	}{
		{
			"StoreInVault with success",
			args{&TxRequest{
				Type:               "sale",
				Amount:             NewDecimal(6500, 2),
				PaymentMethodNonce: FakeNonceVisaCheckoutVisa,
				MerchantAccountId:  testSubMerchantAccount(),
				ServiceFeeAmount:   NewDecimal(1000, 2),
				Options: &TxOpts{
					SubmitForSettlement: true,
					StoreInVault:        true,
				},
			}},
			true,
		},
		{
			"StoreInVault with failure",
			args{&TxRequest{
				Type: "sale",
				// This amount should make the transaction to be declined
				Amount: NewDecimal(200100, 2),
				// This declined nonce is not working in the sandbox
				PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
				MerchantAccountId:  testSubMerchantAccount(),
				ServiceFeeAmount:   NewDecimal(1000, 2),
				Options: &TxOpts{
					SubmitForSettlement: true,
					StoreInVault:        true,
				},
			}},
			true,
		},
		{
			"No StoreInVault",
			args{&TxRequest{
				Type:               "sale",
				Amount:             NewDecimal(6500, 2),
				PaymentMethodNonce: FakeNonceVisaCheckoutVisa,
				MerchantAccountId:  testSubMerchantAccount(),
				ServiceFeeAmount:   NewDecimal(1000, 2),
				Options: &TxOpts{
					SubmitForSettlement: true,
				},
			}},
			false,
		},
		{
			"StoreInVaultOnSuccess with success",
			args{&TxRequest{
				Type:               "sale",
				Amount:             NewDecimal(6500, 2),
				PaymentMethodNonce: FakeNonceVisaCheckoutVisa,
				MerchantAccountId:  testSubMerchantAccount(),
				ServiceFeeAmount:   NewDecimal(1000, 2),
				Options: &TxOpts{
					SubmitForSettlement:   true,
					StoreInVaultOnSuccess: true,
				},
			}},
			true,
		},
		{
			"StoreInVaultOnSuccess with failure",
			args{&TxRequest{
				Type: "sale",
				// This amount should make the transaction to be declined
				Amount: NewDecimal(200100, 2),
				// This declined nonce is not working in the sandbox
				PaymentMethodNonce: FakeNonceProcessorDeclinedVisa,
				MerchantAccountId:  testSubMerchantAccount(),
				ServiceFeeAmount:   NewDecimal(1000, 2),
				Options: &TxOpts{
					SubmitForSettlement:   true,
					StoreInVaultOnSuccess: true,
				},
			}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			txn, err := client.Pay(context.Background(), tt.args.request)

			if err != nil && err.Error() != "Insufficient Funds" {
				t.Fatal(err)
			}

			// Casting transaction from error in order to get the created token
			// in the next checks
			if err != nil && txn == nil {
				txn = err.(*APIError).Transaction
				if txn.Status != StatusProcessorDeclined {
					t.Fatalf("Got status %q, want %q", txn.Status, StatusProcessorDeclined)
				}
			}

			if tt.wantToken &&
				(txn.CreditCard == nil || (txn.CreditCard != nil && txn.CreditCard.Token == "")) {
				t.Error("Success Tx should create token if StoreInVaultOnSuccess equals true")
			}

			if !tt.wantToken && (txn.CreditCard != nil && txn.CreditCard.Token != "") {
				t.Error("Success Tx should NOT create token if StoreInVaultOnSuccess equals false")
			}
		})
	}
}

func TestSettleTransaction(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	prodGateway := New(ProductionURL, "my_merchant_id", "my_public_key", "my_private_key")

	_, err = prodGateway.SandboxSettle(context.Background(), txn.Id)
	if err.Error() != "Operation not allowed in production environment" {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}
}

func TestSettlementConfirmTransaction(t *testing.T) {

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	prodGateway := New(ProductionURL, "my_merchant_id", "my_public_key", "my_private_key")

	_, err = prodGateway.SandboxSettlementConfirm(context.Background(), txn.Id)
	if err.Error() != "Operation not allowed in production environment" {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettlementConfirm(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettlementConfirmed {
		t.Fatal(txn.Status)
	}
}

func TestSettlementDeclinedTransaction(t *testing.T) {

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	prodGateway := New(ProductionURL, "my_merchant_id", "my_public_key", "my_private_key")

	_, err = prodGateway.SandboxSettlementDecline(context.Background(), txn.Id)
	if err.Error() != "Operation not allowed in production environment" {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettlementDecline(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettlementDeclined {
		t.Fatal(txn.Status)
	}
}

func TestSettlementPendingTransaction(t *testing.T) {

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: randomAmount(),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	prodGateway := New(ProductionURL, "my_merchant_id", "my_public_key", "my_private_key")

	_, err = prodGateway.SandboxSettlementPending(context.Background(), txn.Id)
	if err.Error() != "Operation not allowed in production environment" {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettlementPending(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettlementPending {
		t.Fatal(txn.Status)
	}
}

func TestTransactionCreateSettleCheckCreditCardDetails(t *testing.T) {
	t.Parallel()

	amount := NewDecimal(10000, 2)
	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCardDiscover,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if txn.PaymentInstrumentType != CreditCardType {
		t.Fatalf("Returned payment instrument doesn't match input, expected %q, got %q",
			CreditCardType, txn.PaymentInstrumentType)
	}
	if txn.CreditCard.CardType != "Discover" {
		t.Fatalf("Returned credit card detail doesn't match input, expected %q, got %q",
			"Visa", txn.CreditCard.CardType)
	}

	txn, err = client.SubmitForSettlement(context.Background(), txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	txn, err = client.SandboxSettle(context.Background(), txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != StatusSettled {
		t.Fatal(txn.Status)
	}
}

func TestTransactionAndroidPayDetailsAndroidPayProxyCardNonce(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceAndroidPayDiscover,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.AndroidPayDetails == nil {
		t.Fatal("Expected AndroidPayDetail for transaction created with AndroidPay nonce")
	}

	if tx.AndroidPayDetails.CardType == "" {
		t.Fatal("Expected AndroidPayDetail to have CardType set")
	}
	if tx.AndroidPayDetails.Last4 == "" {
		t.Fatal("Expected AndroidPayDetail to have Last4 set")
	}
	if tx.AndroidPayDetails.SourceCardType == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceCardType set")
	}
	if tx.AndroidPayDetails.SourceCardLast4 == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceCardLast4 set")
	}
	if tx.AndroidPayDetails.SourceDescription == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceDescription set")
	}
	if tx.AndroidPayDetails.VirtualCardType == "" {
		t.Fatal("Expected AndroidPayDetail to have VirtualCardType set")
	}
	if tx.AndroidPayDetails.VirtualCardLast4 == "" {
		t.Fatal("Expected AndroidPayDetail to have VirtualCardLast4 set")
	}
	if tx.AndroidPayDetails.ExpirationMonth == "" {
		t.Fatal("Expected AndroidPayDetail to have ExpirationMonth set")
	}
	if tx.AndroidPayDetails.ExpirationYear == "" {
		t.Fatal("Expected AndroidPayDetail to have ExpirationYear set")
	}
	if tx.AndroidPayDetails.GoogleTransactionID == "" {
		t.Fatal("Expected AndroidPayDetail to have GoogleTransactionID set")
	}
	if tx.AndroidPayDetails.BIN == "" {
		t.Fatal("Expected AndroidPayDetail to have BIN set")
	}
}

func TestTransactionAndroidPayDetailsAndroidPayNetworkTokenNonce(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceAndroidPayMasterCard,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.AndroidPayDetails == nil {
		t.Fatal("Expected AndroidPayDetail for transaction created with AndroidPay nonce")
	}

	if tx.AndroidPayDetails.CardType == "" {
		t.Fatal("Expected AndroidPayDetail to have CardType set")
	}
	if tx.AndroidPayDetails.Last4 == "" {
		t.Fatal("Expected AndroidPayDetail to have Last4 set")
	}
	if tx.AndroidPayDetails.SourceCardType == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceCardType set")
	}
	if tx.AndroidPayDetails.SourceCardLast4 == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceCardLast4 set")
	}
	if tx.AndroidPayDetails.SourceDescription == "" {
		t.Fatal("Expected AndroidPayDetail to have SourceDescription set")
	}
	if tx.AndroidPayDetails.VirtualCardType == "" {
		t.Fatal("Expected AndroidPayDetail to have VirtualCardType set")
	}
	if tx.AndroidPayDetails.VirtualCardLast4 == "" {
		t.Fatal("Expected AndroidPayDetail to have VirtualCardLast4 set")
	}
	if tx.AndroidPayDetails.ExpirationMonth == "" {
		t.Fatal("Expected AndroidPayDetail to have ExpirationMonth set")
	}
	if tx.AndroidPayDetails.ExpirationYear == "" {
		t.Fatal("Expected AndroidPayDetail to have ExpirationYear set")
	}
	if tx.AndroidPayDetails.GoogleTransactionID == "" {
		t.Fatal("Expected AndroidPayDetail to have GoogleTransactionID set")
	}
	if tx.AndroidPayDetails.BIN == "" {
		t.Fatal("Expected AndroidPayDetail to have BIN set")
	}
}

func TestTransactionWithoutAndroidPayDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceTransactable,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.AndroidPayDetails != nil {
		t.Fatalf("Expected AndroidPayDetail to be nil for transaction created without AndroidPay, but was %#v", tx.AndroidPayDetails)
	}
}

func TestTransactionApplePayDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceApplePayVisa,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.ApplePayDetails == nil {
		t.Fatal("Expected ApplePayDetail for transaction created with ApplePay nonce")
	}

	wantNonceCardType := "Apple Pay - Visa"
	if tx.ApplePayDetails.CardType != wantNonceCardType {
		t.Errorf("Got ApplePayDetail.CardType %v, want %v", tx.ApplePayDetails.CardType, wantNonceCardType)
	}
	if tx.ApplePayDetails.PaymentInstrumentName == "" {
		t.Fatal("Expected ApplePayDetail to have PaymentInstrumentName set")
	}
	if tx.ApplePayDetails.SourceDescription == "" {
		t.Fatal("Expected ApplePayDetail to have SourceDescription set")
	}
	if tx.ApplePayDetails.CardholderName == "" {
		t.Fatal("Expected ApplePayDetail to have CardholderName set")
	}
	if !ValidExpiryMonth(tx.ApplePayDetails.ExpirationMonth) {
		t.Errorf("ApplePayDetail.ExpirationMonth (%s) does not match expected value", tx.ApplePayDetails.ExpirationMonth)
	}
	if !ValidExpiryYear(tx.ApplePayDetails.ExpirationYear) {
		t.Errorf("ApplePayDetail.ExpirationYear (%s) does not match expected value", tx.ApplePayDetails.ExpirationYear)
	}
	if !ValidBIN(tx.ApplePayDetails.BIN) {
		t.Errorf("ApplePayDetail.BIN (%s) does not conform expected value", tx.ApplePayDetails.BIN)
	}
	if !ValidLast4(tx.ApplePayDetails.Last4) {
		t.Errorf("ApplePayDetail.Last4 (%s) does not conform match value", tx.ApplePayDetails.Last4)
	}
}

func TestTransactionWithoutApplePayDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceTransactable,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.ApplePayDetails != nil {
		t.Fatalf("Expected ApplePayDetail to be nil for transaction created without ApplePay, but was %#v", tx.ApplePayDetails)
	}
}

func TestTransactionClone(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(2000, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	// Clone
	tx2, err := client.Clone(context.Background(), tx.Id, &TxCloneRequest{
		Amount:  NewDecimal(1000, 2),
		Channel: "ChannelA",
		Options: &TxCloneOpts{
			SubmitForSettlement: false,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if g, w := tx2.Status, StatusAuthorized; g != w {
		t.Errorf("Tx status got %v, want %v", g, w)
	}
	if g, w := tx2.Amount, NewDecimal(1000, 2); g.Cmp(w) != 0 {
		t.Errorf("Tx amount got %v, want %v", g, w)
	}
	if g, w := tx2.Channel, "ChannelA"; g != w {
		t.Errorf("Tx channel got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.ExpirationMonth, "05"; g != w {
		t.Errorf("Tx credit card expiration month got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.ExpirationYear, "2014"; g != w {
		t.Errorf("Tx credit card expiration year got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.Last4, "1111"; g != w {
		t.Errorf("Tx credit card last 4 got %v, want %v", g, w)
	}
}

func TestTransactionCloneSubmittedForSettlement(t *testing.T) {
	t.Parallel()

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:   "sale",
		Amount: NewDecimal(2000, 2),
		CreditCard: &CreditCard{
			Number:         testCardVisa,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Clone
	tx2, err := client.Clone(context.Background(), tx.Id, &TxCloneRequest{
		Amount:  NewDecimal(1000, 2),
		Channel: "ChannelA",
		Options: &TxCloneOpts{
			SubmitForSettlement: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if g, w := tx2.Status, StatusSubmittedForSettlement; g != w {
		t.Errorf("Tx status got %v, want %v", g, w)
	}
	if g, w := tx2.Amount, NewDecimal(1000, 2); g.Cmp(w) != 0 {
		t.Errorf("Tx amount got %v, want %v", g, w)
	}
	if g, w := tx2.Channel, "ChannelA"; g != w {
		t.Errorf("Tx channel got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.ExpirationMonth, "05"; g != w {
		t.Errorf("Tx credit card expiration month got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.ExpirationYear, "2014"; g != w {
		t.Errorf("Tx credit card expiration year got %v, want %v", g, w)
	}
	if g, w := tx2.CreditCard.Last4, "1111"; g != w {
		t.Errorf("Tx credit card last 4 got %v, want %v", g, w)
	}
}

func TestTransactionPayPalDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNoncePayPalOneTimePayment,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.PayPalDetails == nil {
		t.Fatal("Expected PayPalDetail for transaction created with PayPal nonce")
	}

	if tx.PayPalDetails.PayerEmail == "" {
		t.Fatal("Expected PayPalDetail to have PayerEmail set")
	}
	if tx.PayPalDetails.PaymentID == "" {
		t.Fatal("Expected PayPalDetail to have PaymentID set")
	}
	if tx.PayPalDetails.ImageURL == "" {
		t.Fatal("Expected PayPalDetail to have ImageURL set")
	}
	if tx.PayPalDetails.DebugID == "" {
		t.Fatal("Expected PayPalDetail to have DebugID set")
	}
	if tx.PayPalDetails.PayerID == "" {
		t.Fatal("Expected PayPalDetail to have DebugID set")
	}
	if tx.PayPalDetails.PayerFirstName == "" {
		t.Fatal("Expected PayPalDetail to have PayerFirstName set")
	}
	if tx.PayPalDetails.PayerLastName == "" {
		t.Fatal("Expected PayPalDetail to have PayerLastName set")
	}
	if tx.PayPalDetails.PayerStatus == "" {
		t.Fatal("Expected PayPalDetail to have PayerStatus set")
	}
	if tx.PayPalDetails.SellerProtectionStatus == "" {
		t.Fatal("Expected PayPalDetail to have SellerProtectionStatus set")
	}
}

func TestTransactionWithoutPayPalDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceTransactable,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.PayPalDetails != nil {
		t.Fatalf("Expected PayPalDetail to be nil for transaction created without PayPal, but was %#v", tx.PayPalDetails)
	}
}

func TestTransactionVenmoAccountDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceVenmoAccount,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.VenmoAccountDetails == nil {
		t.Fatal("Expected VenmoAccountDetail for transaction created with VenmoAccount nonce")
	}

	if tx.VenmoAccountDetails.Token != "" {
		t.Fatal("Expected VenmoAccountDetail to not have Token set")
	}
	if tx.VenmoAccountDetails.Username == "" {
		t.Fatal("Expected VenmoAccountDetail to have Username set")
	}
	if tx.VenmoAccountDetails.VenmoUserID == "" {
		t.Fatal("Expected VenmoAccountDetail to have VenmoUserID set")
	}
	if tx.VenmoAccountDetails.SourceDescription == "" {
		t.Fatal("Expected VenmoAccountDetail to have SourceDescription set")
	}
	if tx.VenmoAccountDetails.ImageURL == "" {
		t.Fatal("Expected VenmoAccountDetail to have ImageURL set")
	}
}

func TestTransactionWithoutVenmoAccountDetails(t *testing.T) {

	tx, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(2000, 2),
		PaymentMethodNonce: FakeNonceTransactable,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != StatusAuthorized {
		t.Fatal(tx.Status)
	}

	if tx.VenmoAccountDetails != nil {
		t.Fatalf("Expected VenmoAccountDetail to be nil for transaction created without a VenmoAccount, but was %#v", tx.VenmoAccountDetails)
	}
}

func TestTransactionWithLineItemsZero(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
	})

	if err != nil {
		t.Fatal(err)
	}

	lineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(lineItems), 0; g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}
}

func TestTransactionWithLineItemsSingleOnlyRequiredFields(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:        "Name #1",
				Kind:        LineItemDebitKind,
				Quantity:    NewDecimal(10232, 4),
				UnitAmount:  NewDecimal(451232, 4),
				TotalAmount: NewDecimal(4515, 2),
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	lineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(lineItems), 1; g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}

	l := lineItems[0]
	if g, w := l.Name, "Name #1"; g != w {
		t.Errorf("got name %q, want %q", g, w)
	}
	if g, w := l.Kind, LineItemDebitKind; g != w {
		t.Errorf("got kind %q, want %q", g, w)
	}
	if g, w := l.Quantity, NewDecimal(10232, 4); g.Cmp(w) != 0 {
		t.Errorf("got quantity %q, want %q", g, w)
	}
	if g, w := l.UnitAmount, NewDecimal(451232, 4); g.Cmp(w) != 0 {
		t.Errorf("got unit amount %q, want %q", g, w)
	}
	if g, w := l.TotalAmount, NewDecimal(4515, 2); g.Cmp(w) != 0 {
		t.Errorf("got total amount %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsSingleZeroAmountFields(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(0, 0),
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(0, 0),
				DiscountAmount: NewDecimal(0, 0),
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	lineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(lineItems), 1; g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}

	l := lineItems[0]
	if g, w := l.Name, "Name #1"; g != w {
		t.Errorf("got name %q, want %q", g, w)
	}
	if g, w := l.Kind, LineItemDebitKind; g != w {
		t.Errorf("got kind %q, want %q", g, w)
	}
	if g, w := l.Quantity, NewDecimal(10232, 4); g.Cmp(w) != 0 {
		t.Errorf("got quantity %q, want %q", g, w)
	}
	if g, w := l.UnitAmount, NewDecimal(451232, 4); g.Cmp(w) != 0 {
		t.Errorf("got unit amount %q, want %q", g, w)
	}
	if g, w := l.UnitTaxAmount, NewDecimal(0, 0); g.Cmp(w) != 0 {
		t.Errorf("got unit tax amount %q, want %q", g, w)
	}
	if g, w := l.TotalAmount, NewDecimal(4515, 2); g.Cmp(w) != 0 {
		t.Errorf("got total amount %q, want %q", g, w)
	}
	if g, w := l.TaxAmount, NewDecimal(0, 0); g.Cmp(w) != 0 {
		t.Errorf("got tax amount %q, want %q", g, w)
	}
	if g, w := l.DiscountAmount, NewDecimal(0, 0); g.Cmp(w) != 0 {
		t.Errorf("got discount amount %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsSingle(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	lineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(lineItems), 1; g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}

	l := lineItems[0]
	if g, w := l.Name, "Name #1"; g != w {
		t.Errorf("got name %q, want %q", g, w)
	}
	if g, w := l.Description, "Description #1"; g != w {
		t.Errorf("got description %q, want %q", g, w)
	}
	if g, w := l.Kind, LineItemDebitKind; g != w {
		t.Errorf("got kind %q, want %q", g, w)
	}
	if g, w := l.Quantity, NewDecimal(10232, 4); g.Cmp(w) != 0 {
		t.Errorf("got quantity %q, want %q", g, w)
	}
	if g, w := l.UnitAmount, NewDecimal(451232, 4); g.Cmp(w) != 0 {
		t.Errorf("got unit amount %q, want %q", g, w)
	}
	if g, w := l.UnitTaxAmount, NewDecimal(123, 2); g.Cmp(w) != 0 {
		t.Errorf("got unit tax amount %q, want %q", g, w)
	}
	if g, w := l.UnitOfMeasure, "gallon"; g != w {
		t.Errorf("got unit of measure %q, want %q", g, w)
	}
	if g, w := l.TotalAmount, NewDecimal(4515, 2); g.Cmp(w) != 0 {
		t.Errorf("got total amount %q, want %q", g, w)
	}
	if g, w := l.TaxAmount, NewDecimal(455, 2); g.Cmp(w) != 0 {
		t.Errorf("got tax amount %q, want %q", g, w)
	}
	if g, w := l.DiscountAmount, NewDecimal(102, 2); g.Cmp(w) != 0 {
		t.Errorf("got discount amount %q, want %q", g, w)
	}
	if g, w := l.ProductCode, "23434"; g != w {
		t.Errorf("got product code %q, want %q", g, w)
	}
	if g, w := l.CommodityCode, "9SAASSD8724"; g != w {
		t.Errorf("got commodity code %q, want %q", g, w)
	}
	if g, w := l.URL, "https://example.com/products/23434"; g != w {
		t.Errorf("got url %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsMultiple(t *testing.T) {
	t.Parallel()

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:          "Name #2",
				Description:   "Description #2",
				Kind:          LineItemCreditKind,
				Quantity:      NewDecimal(202, 2),
				UnitAmount:    NewDecimal(5, 0),
				UnitOfMeasure: "gallon",
				TotalAmount:   NewDecimal(4515, 2),
				TaxAmount:     NewDecimal(455, 2),
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	lineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(lineItems), 2; g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}

	{
		l := lineItems[0]
		if g, w := l.Name, "Name #1"; g != w {
			t.Errorf("got name %q, want %q", g, w)
		}
		if g, w := l.Description, "Description #1"; g != w {
			t.Errorf("got description %q, want %q", g, w)
		}
		if g, w := l.Kind, LineItemDebitKind; g != w {
			t.Errorf("got kind %q, want %q", g, w)
		}
		if g, w := l.Quantity, NewDecimal(10232, 4); g.Cmp(w) != 0 {
			t.Errorf("got quantity %q, want %q", g, w)
		}
		if g, w := l.UnitAmount, NewDecimal(451232, 4); g.Cmp(w) != 0 {
			t.Errorf("got unit amount %q, want %q", g, w)
		}
		if g, w := l.UnitTaxAmount, NewDecimal(123, 2); g.Cmp(w) != 0 {
			t.Errorf("got unit tax amount %q, want %q", g, w)
		}
		if g, w := l.UnitOfMeasure, "gallon"; g != w {
			t.Errorf("got unit of measure %q, want %q", g, w)
		}
		if g, w := l.TotalAmount, NewDecimal(4515, 2); g.Cmp(w) != 0 {
			t.Errorf("got total amount %q, want %q", g, w)
		}
		if g, w := l.TaxAmount, NewDecimal(455, 2); g.Cmp(w) != 0 {
			t.Errorf("got tax amount %q, want %q", g, w)
		}
		if g, w := l.DiscountAmount, NewDecimal(102, 2); g.Cmp(w) != 0 {
			t.Errorf("got discount amount %q, want %q", g, w)
		}
		if g, w := l.ProductCode, "23434"; g != w {
			t.Errorf("got product code %q, want %q", g, w)
		}
		if g, w := l.CommodityCode, "9SAASSD8724"; g != w {
			t.Errorf("got commodity code %q, want %q", g, w)
		}
		if g, w := l.URL, "https://example.com/products/23434"; g != w {
			t.Errorf("got url %q, want %q", g, w)
		}
	}

	{
		l := lineItems[1]
		if g, w := l.Name, "Name #2"; g != w {
			t.Errorf("got name %q, want %q", g, w)
		}
		if g, w := l.Description, "Description #2"; g != w {
			t.Errorf("got description %q, want %q", g, w)
		}
		if g, w := l.Kind, LineItemCreditKind; g != w {
			t.Errorf("got kind %q, want %q", g, w)
		}
		if g, w := l.Quantity, NewDecimal(202, 2); g.Cmp(w) != 0 {
			t.Errorf("got quantity %q, want %q", g, w)
		}
		if g, w := l.UnitAmount, NewDecimal(5, 0); g.Cmp(w) != 0 {
			t.Errorf("got unit amount %q, want %q", g, w)
		}
		if g, w := l.UnitOfMeasure, "gallon"; g != w {
			t.Errorf("got unit of measure %q, want %q", g, w)
		}
		if g, w := l.TotalAmount, NewDecimal(4515, 2); g.Cmp(w) != 0 {
			t.Errorf("got total amount %q, want %q", g, w)
		}
		if g, w := l.TaxAmount, NewDecimal(455, 2); g.Cmp(w) != 0 {
			t.Errorf("got tax amount %q, want %q", g, w)
		}
		if g, w := l.DiscountAmount, (*Decimal)(nil); g != nil {
			t.Errorf("got discount amount %q, want %q", g, w)
		}
		if g, w := l.ProductCode, ""; g != w {
			t.Errorf("got product code %q, want %q", g, w)
		}
		if g, w := l.CommodityCode, ""; g != w {
			t.Errorf("got commodity code %q, want %q", g, w)
		}
		if g, w := l.URL, ""; g != w {
			t.Errorf("got url %q, want %q", g, w)
		}
	}
}

func TestTransactionWithLineItemsMaxMultiple(t *testing.T) {
	t.Parallel()

	lineItems := LineItemRequests{}
	for i := 0; i < 249; i++ {
		lineItems = append(lineItems, &LineItemRequest{
			Name:           "Name #1",
			Description:    "Description #1",
			Kind:           LineItemDebitKind,
			Quantity:       NewDecimal(10232, 4),
			UnitAmount:     NewDecimal(451232, 4),
			UnitTaxAmount:  NewDecimal(123, 2),
			UnitOfMeasure:  "gallon",
			TotalAmount:    NewDecimal(4515, 2),
			TaxAmount:      NewDecimal(455, 2),
			DiscountAmount: NewDecimal(102, 2),
			ProductCode:    "23434",
			CommodityCode:  "9SAASSD8724",
			URL:            "https://example.com/products/23434",
		})
	}

	txn, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems:          lineItems,
	})

	if err != nil {
		t.Fatal(err)
	}

	foundLineItems, err := client.FindTransactionLineItem(context.Background(), txn.Id)

	if err != nil {
		t.Fatal(err)
	}

	if g, w := len(foundLineItems), len(lineItems); g != w {
		t.Fatalf("got %d line items, want %d line items", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorCommodityCodeIsTooLong(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "0123456789123",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("CommodityCode")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95801"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "CommodityCode"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Commodity code is too long."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorDescriptionIsTooLong(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "This is a line item description which is far too long. Like, way too long to be practical. We don't like how long this line item description is.",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Description")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95803"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Description"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Description is too long."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorDiscountAmountIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(214748364800, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("DiscountAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95805"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "DiscountAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Discount amount is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorDiscountAmountCannotBeNegative(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(-200, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("DiscountAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95806"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "DiscountAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Discount amount cannot be negative."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTaxAmountIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(214748364800, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("TaxAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95828"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "TaxAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Tax amount is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTaxAmountCannotBeNegative(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(-200, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("TaxAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95829"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "TaxAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Tax amount cannot be negative."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorKindIsRequired(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Kind")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95808"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Kind"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Kind is required."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorNameIsRequired(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Name")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95822"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Name"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Name is required."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorNameIsTooLong(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "123456789012345678901234567890123456",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Name")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95823"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Name"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Name is too long."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorProductCodeIsTooLong(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "123456789012345678901234567890123456",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("ProductCode")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95809"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "ProductCode"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Product code is too long."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorQuantityIsRequired(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Quantity")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95811"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Quantity"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Quantity is required."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorQuantityIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(21474836480000, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("Quantity")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95812"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "Quantity"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Quantity is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTotalAmountIsRequired(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("TotalAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95814"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "TotalAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Total amount is required."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTotalAmountIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(214748364800, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("TotalAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95815"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "TotalAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Total amount is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTotalAmountMustBeGreaterThanZero(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(-200, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("TotalAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95816"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "TotalAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Total amount must be greater than zero."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitAmountIsRequired(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95818"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit amount is required."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitAmountIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(21474836480000, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95819"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit amount is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitAmountMustBeGreaterThanZero(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(-20000, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95820"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit amount must be greater than zero."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitOfMeasureIsTooLong(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "1234567890123",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitOfMeasure")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95821"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitOfMeasure"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit of measure is too long."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitTaxAmountIsInvalid(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(1234, 3),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitTaxAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95824"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitTaxAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit tax amount is an invalid format."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitTaxAmountIsTooLarge(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(214748364800, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitTaxAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95825"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitTaxAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit tax amount is too large."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorUnitTaxAmountCannotBeNegative(t *testing.T) {
	t.Parallel()

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems: LineItemRequests{
			&LineItemRequest{
				Name:           "Name #1",
				Description:    "Description #1",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(123, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
			&LineItemRequest{
				Name:           "Name #2",
				Description:    "Description #2",
				Kind:           LineItemDebitKind,
				Quantity:       NewDecimal(10232, 4),
				UnitAmount:     NewDecimal(451232, 4),
				UnitTaxAmount:  NewDecimal(-200, 2),
				UnitOfMeasure:  "gallon",
				TotalAmount:    NewDecimal(4515, 2),
				TaxAmount:      NewDecimal(455, 2),
				DiscountAmount: NewDecimal(102, 2),
				ProductCode:    "23434",
				CommodityCode:  "9SAASSD8724",
				URL:            "https://example.com/products/23434",
			},
		},
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").For("LineItems").ForIndex(1).On("UnitTaxAmount")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "95826"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "UnitTaxAmount"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Unit tax amount cannot be negative."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}

func TestTransactionWithLineItemsValidationErrorTooManyLineItems(t *testing.T) {
	t.Parallel()

	lineItems := LineItemRequests{}
	for i := 0; i < 250; i++ {
		lineItems = append(lineItems, &LineItemRequest{
			Name:           "Name #1",
			Description:    "Description #1",
			Kind:           LineItemDebitKind,
			Quantity:       NewDecimal(10232, 4),
			UnitAmount:     NewDecimal(451232, 4),
			UnitTaxAmount:  NewDecimal(123, 2),
			UnitOfMeasure:  "gallon",
			TotalAmount:    NewDecimal(4515, 2),
			TaxAmount:      NewDecimal(455, 2),
			DiscountAmount: NewDecimal(102, 2),
			ProductCode:    "23434",
			CommodityCode:  "9SAASSD8724",
			URL:            "https://example.com/products/23434",
		})
	}

	_, err := client.Pay(context.Background(), &TxRequest{
		Type:               "sale",
		Amount:             NewDecimal(1423, 2),
		PaymentMethodNonce: FakeNonceTransactable,
		LineItems:          lineItems,
	})

	if err == nil {
		t.Fatal("got no error, want error")
	}

	allValidationErrors := err.(*APIError).All()
	if g, w := len(allValidationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}

	validationErrors := err.(*APIError).For("Transaction").On("LineItems")
	if g, w := len(validationErrors), 1; g != w {
		t.Errorf("got %d errors, want %d", g, w)
	}
	if g, w := validationErrors[0].Code, "915157"; g != w {
		t.Errorf("got error code %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Attribute, "LineItems"; g != w {
		t.Errorf("got error attribute %q, want %q", g, w)
	}
	if g, w := validationErrors[0].Message, "Too many line items."; g != w {
		t.Errorf("got error message %q, want %q", g, w)
	}
}
