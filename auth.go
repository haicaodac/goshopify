package goshopify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	userAgent             = "goshopify/1.0.0"
	shopifyChecksumHeader = "X-Shopify-Hmac-Sha256"
)

// Auth ...
func (app *AppShopify) Auth(shopName string) (string, error) {
	shopURL, _ := url.Parse(ShopBaseURL(shopName))
	shopURL.Path = "/admin/oauth/authorize"
	query := shopURL.Query()
	query.Set("client_id", app.config.APIKey)
	query.Set("redirect_uri", app.config.RedirectUrl)
	query.Set("scope", app.config.Scope)
	// query.Set("state", state)
	shopURL.RawQuery = query.Encode()
	return shopURL.String(), nil
}

// VerifyAuthURL ... Verifying URL callback parameters.
func (app *AppShopify) VerifyAuthURL(u *url.URL) (bool, error) {
	q := u.Query()
	messageMAC := q.Get("hmac")

	// Remove hmac and signature and leave the rest of the parameters alone.
	q.Del("hmac")
	q.Del("signature")

	message, err := url.QueryUnescape(q.Encode())

	return app.VerifyMessage(message, messageMAC), err
}

//VerifyMessage ... Verify a message against a message HMAC
func (app *AppShopify) VerifyMessage(message, messageMAC string) bool {
	mac := hmac.New(sha256.New, []byte(app.config.APISecret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	// shopify HMAC is in hex so it needs to be decoded
	actualMac, _ := hex.DecodeString(messageMAC)

	return hmac.Equal(actualMac, expectedMAC)
}

// GetAccessToken ...
func (app *AppShopify) GetAccessToken(shopName string, code string) (string, error) {
	data := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientID:     app.config.APIKey,
		ClientSecret: app.config.APISecret,
		Code:         code,
	}

	client := NewClient(*app, shopName, "")

	rel, err := url.Parse("admin/oauth/access_token")
	if err != nil {
		return "", err
	}
	// Make the full url based on the relative path
	url := client.baseURL.ResolveReference(rel)

	var js []byte
	js, _ = json.Marshal(data)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(js))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", userAgent)

	clientReq := &http.Client{}
	resp, _ := clientReq.Do(req)

	type Token struct {
		Token string `json:"access_token"`
	}
	token := Token{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		return "", err
	}

	client.token = token.Token

	return token.Token, err
}

// VerifyWebhookRequest ... Verifies a webhook http request, sent by Shopify.
// The body of the request is still readable after invoking the method.
func (app *AppShopify) VerifyWebhookRequest(httpRequest *http.Request) bool {
	shopifySha256 := httpRequest.Header.Get(shopifyChecksumHeader)
	actualMac := []byte(shopifySha256)

	mac := hmac.New(sha256.New, []byte(app.config.APISecret))
	requestBody, _ := ioutil.ReadAll(httpRequest.Body)
	httpRequest.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	mac.Write(requestBody)
	macSum := mac.Sum(nil)
	expectedMac := []byte(base64.StdEncoding.EncodeToString(macSum))

	return hmac.Equal(actualMac, expectedMac)
}

// Verifies a webhook http request, sent by Shopify.
// The body of the request is still readable after invoking the method.
// This method has more verbose error output which is useful for debugging.
func (app *AppShopify) VerifyWebhookRequestVerbose(httpRequest *http.Request) (bool, error) {
	if app.config.APISecret == "" {
		return false, errors.New("ApiSecret is empty")
	}

	shopifySha256 := httpRequest.Header.Get(shopifyChecksumHeader)
	if shopifySha256 == "" {
		return false, fmt.Errorf("header %s not set", shopifyChecksumHeader)
	}

	decodedReceivedHMAC, err := base64.StdEncoding.DecodeString(shopifySha256)
	if err != nil {
		return false, err
	}
	if len(decodedReceivedHMAC) != 32 {
		return false, fmt.Errorf("received HMAC is not of length 32, it is of length %d", len(decodedReceivedHMAC))
	}

	mac := hmac.New(sha256.New, []byte(app.config.APISecret))
	requestBody, err := ioutil.ReadAll(httpRequest.Body)
	if err != nil {
		return false, err
	}

	httpRequest.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	if len(requestBody) == 0 {
		return false, errors.New("request body is empty")
	}

	// Sha256 write doesn't actually return an error
	mac.Write(requestBody)

	computedHMAC := mac.Sum(nil)
	HMACSame := hmac.Equal(decodedReceivedHMAC, computedHMAC)
	if !HMACSame {
		return HMACSame, fmt.Errorf("expected hash %x does not equal %x", computedHMAC, decodedReceivedHMAC)
	}

	return HMACSame, nil
}
