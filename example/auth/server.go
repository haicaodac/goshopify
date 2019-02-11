package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/haicaodac/goshopify"
)

var app goshopify.AppShopify

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func main() {
	app = goshopify.New(goshopify.Config{
		APIKey:      "53dfe43ed6e0e8f86ab6dadb75e8ed81",
		APISecret:   "40a9f6ee2c06395a401fd3a2258a566a",
		RedirectUrl: "http://localhost:8080/shopify/callback",
		Scope:       "read_products,read_orders",
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cookie-name")

		// Revoke users authentication
		session.Values["authenticated"] = false
		session.Save(r, w)
	})

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cookie-name")

		// Check if user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

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
			session.Values["authenticated"] = true
			session.Save(r, w)
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
		query := r.URL.Query()
		shopName := query.Get("shop")
		code := query.Get("code")

		// Check that the callback signature is valid
		if ok, _ := app.VerifyAuthURL(r.URL); !ok {
			http.Error(w, "Invalid Signature", http.StatusUnauthorized)
			return
		}
		// url := "http://localhost:8080/shopify/callback?code=581c1ae239749d5d851bc6b4b305b80a&hmac=94d58eb6fe6c3cddaf367de02a994153cd45f6c5d4bc2107fb3180f4455702b8&shop=dachai.myshopify.com&timestamp=1548826544"
		token, err := app.GetAccessToken(shopName, code)
		url := "http://localhost:8080/dashboard"
		// log.Println(url)

		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = true
		session.Save(r, w)

		http.Redirect(w, r, url, http.StatusFound)

		fmt.Println("-------------------------------token-----------------------------")
		fmt.Println(token)
		fmt.Println(err)

	})

	fmt.Println("Server run ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
