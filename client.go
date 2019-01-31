package goshopify

import (
	"net/http"
	"net/url"
)

// Client manages communication with the Shopify API.
type Client struct {
	// HTTP client used to communicate with the DO API.
	Client *http.Client

	// App settings
	app AppShopify

	// Base URL for API requests.
	// This is set on a per-store basis which means that each store must have
	// its own client.
	baseURL *url.URL

	// A permanent access token
	token string
}

// NewClient ...
func NewClient(app AppShopify, shopName, token string) *Client {
	httpClient := http.DefaultClient

	baseURL, _ := url.Parse(ShopBaseURL(shopName))

	c := &Client{Client: httpClient, app: app, baseURL: baseURL, token: token}
	return c
}
