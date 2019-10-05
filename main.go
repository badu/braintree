package braintree

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	HdrContentEncoding = "Content-Encoding"
	HdrContentType     = "Content-Type"
	HdrAccept          = "Accept"
	HdrAcceptEncoding  = "Accept-Encoding"
	HdrUserAgent       = "User-Agent"
	HdrAuthorization   = "Authorization"
	HdrApplicationXML  = "application/xml"
)

type apiVersion int

const (
	apiVersion3 apiVersion = 3
	apiVersion4            = 4
)

const defaultTimeout = time.Second * 60
const DateFormat = "2006-01-02"

var (
	defaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	DefaultClient = &http.Client{
		Timeout:   defaultTimeout,
		Transport: defaultTransport,
	}
)

const (
	DevURL        = "http://localhost:3000"
	SandboxURL    = "https://api.sandbox.braintreegateway.com:443"
	ProductionURL = "https://api.braintreegateway.com:443"
)

func SetURLForEnv(name string) (string, error) {
	switch name {
	case "dev":
		return DevURL, nil
	case "sandbox":
		return SandboxURL, nil
	case "prod":
		return ProductionURL, nil
	}
	return "", fmt.Errorf("bad environment %q", name)
}

func New(url, merchId, pubKey, privKey string) *APIClient {
	return NewWithHttpClient(url, merchId, pubKey, privKey, DefaultClient)
}

func NewAPIKey(url string, merchId, publicKey, privateKey string) *Secret {
	return &Secret{
		URL:     url,
		MerchId: merchId,
		Key:     Key{PublicKey: publicKey, PrivateKey: privateKey},
	}
}

func NewWithHttpClient(url string, merchantId, publicKey, privateKey string, client *http.Client) *APIClient {
	return &APIClient{Key: NewAPIKey(url, merchantId, publicKey, privateKey), Client: client}
}

type Secret struct {
	Key
	Raw     string
	URL     string
	MerchId string
}

func NewAccessToken(accessTokenStr string) (*Secret, error) {
	parts := strings.Split(accessTokenStr, "$")
	if len(parts) < 3 || parts[0] != "access_token" {
		return nil, errors.New("access token is not of expected format")
	}
	env, err := SetURLForEnv(parts[1])
	if err != nil {
		return nil, errors.New("access token is for unsupported environment, " + err.Error())
	}
	t := Secret{
		Raw:     accessTokenStr,
		URL:     env,
		MerchId: parts[2],
	}
	return &t, nil
}

func (t Secret) AuthorizationHeader() string {
	if t.Raw == "" {
		//it's an api key, not a bearer header
		auth := t.PublicKey + ":" + t.PrivateKey
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}
	return "Bearer " + t.Raw
}

func NewWithAccessToken(accessToken string) (*APIClient, error) {
	token, err := NewAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	return &APIClient{Key: token, Client: DefaultClient}, nil
}

type APIClient struct {
	Key    *Secret
	Logger *log.Logger
	Client *http.Client
}

func (c *APIClient) do(ctx context.Context, method, path string, payload interface{}) (*Response, error) {
	return c.call(ctx, method, path, payload, apiVersion3)
}

func (c *APIClient) call(ctx context.Context, method, path string, payload interface{}, v apiVersion) (*Response, error) {
	var buf bytes.Buffer
	if payload != nil {
		xmlBody, err := xml.Marshal(payload)
		if err != nil {
			c.Logger.Printf("xml Marshal error : %v\n", err)
			return nil, err
		}
		_, err = buf.Write(xmlBody)
		if err != nil {
			c.Logger.Printf("write error : %v\n", err)
			return nil, err
		}
	}

	url := c.Key.URL + "/merchants/" + c.Key.MerchId + "/" + path

	if c.Logger != nil {
		c.Logger.Printf("---\n%q on %s payload:\n%s\n---\n", method, url, buf.String())
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		c.Logger.Printf("request building error : %v", err)
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set(HdrUserAgent, "Braintree-API-Client")
	req.Header.Set(HdrContentType, HdrApplicationXML)
	req.Header.Set(HdrAccept, HdrApplicationXML)
	req.Header.Set(HdrAcceptEncoding, "gzip")
	req.Header.Set("X-ApiVersion", fmt.Sprintf("%d", v))
	req.Header.Set(HdrAuthorization, c.Key.AuthorizationHeader())

	httpClient := c.Client
	if httpClient == nil {
		httpClient = DefaultClient
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	result := &Response{
		Response: response,
	}
	err = result.unpackBody()
	if err != nil {
		return nil, err
	}

	if c.Logger != nil {
		c.Logger.Printf("Braintree Response: <\n%s", string(result.Body))
	}

	err = result.apiError()
	if err != nil {
		c.Logger.Printf("xml Unmarshal error : %v\n", err)
		return nil, err
	}
	return result, nil
}
