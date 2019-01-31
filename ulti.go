package goshopify

import (
	"fmt"
	"strings"
)

// ShopFullName ... Return the full shop name, including .myshopify.com
func ShopFullName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	if strings.Contains(name, "myshopify.com") {
		return name
	}
	return name + ".myshopify.com"
}

// ShopShortName ... Return the short shop name, excluding .myshopify.com
func ShopShortName(name string) string {
	// Convert to fullname and remove the myshopify part. Perhaps not the most
	// performant solution, but then we don't have to repeat all the trims here
	// :-)pas
	return strings.Replace(ShopFullName(name), ".myshopify.com", "", -1)
}

// ShopBaseURL ... Return the Shop's base url.
func ShopBaseURL(name string) string {
	name = ShopFullName(name)
	return fmt.Sprintf("https://%s", name)
}
