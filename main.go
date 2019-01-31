package goshopify // goshopify

// AppShopify ...
type AppShopify struct {
	config Config
}

var appShopify AppShopify

// New ...
func New(config Config) AppShopify {
	appShopify = AppShopify{
		config,
	}
	return appShopify
}
