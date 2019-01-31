package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/haicaodac/goshopify"
)

var app goshopify.AppShopify

func main() {
	app = goshopify.New(goshopify.Config{
		APIKey:      "7073a9eff6296c1ab69ab9dc546344d1",
		APISecret:   "9f6d680d026f863018e46289b5f30a46",
		RedirectUrl: "http://localhost:8080/shopify/callback",
		Scope:       "read_products,read_orders",
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		shopName := r.URL.Query().Get("shop")
		hMac := r.URL.Query().Get("hmac")

		if shopName != "" && hMac != "" {
			ok, err := app.VerifyAuthURL(r.URL)
			if err != nil {
				http.Error(w, "Invalid Signature", http.StatusUnauthorized)
				return
			}
			if !ok {
				http.Error(w, "Invalid Signature", http.StatusUnauthorized)
				return
			}
			fmt.Println("done")
		} else {
			// Home pages

		}
		fmt.Println(shopName)
		// fmt.Println(authURL)
	})

	http.HandleFunc("/shopify/auth", func(w http.ResponseWriter, r *http.Request) {
		shopName := r.URL.Query().Get("shop")
		// fmt.Println(shopName)
		authURL, _ := app.Auth(shopName)
		// fmt.Println(authURL)
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	http.HandleFunc("/shopify/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check that the callback signature is valid
		if ok, _ := app.VerifyAuthURL(r.URL); !ok {
			http.Error(w, "Invalid Signature", http.StatusUnauthorized)
			return
		}
		// url := "http://localhost:8080/shopify/callback?code=581c1ae239749d5d851bc6b4b305b80a&hmac=94d58eb6fe6c3cddaf367de02a994153cd45f6c5d4bc2107fb3180f4455702b8&shop=dachai.myshopify.com&timestamp=1548826544"
		query := r.URL.Query()
		shopName := query.Get("shop")
		code := query.Get("code")

		token, err := app.GetAccessToken(shopName, code)

		fmt.Println("-------------------------------token-----------------------------")
		fmt.Println(token)
		fmt.Println(err)

	})

	fmt.Println("Server run ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
