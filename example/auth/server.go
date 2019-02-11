package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/haicaodac/goshopify"
)

var app goshopify.AppShopify

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

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
			// Check if user is authenticated
			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		fmt.Println(shopName)
		// fmt.Println(authURL)
	})

	http.HandleFunc("/archive-zip", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("temp"); os.IsNotExist(err) {
			os.Mkdir("temp", 0777)
		}

		html := []byte(`<html>
		<head>
			<title>Xin chao</title>
		</head>
		<body>
			<h1>dsahkjdhasjk</h1>
		</body>
	</html>`)
		err := ioutil.WriteFile("temp/html.liquid", html, 0644)
		if err != nil {
			panic(err)
		}

		css := []byte(`.h1{
			color: red;
		}`)
		err = ioutil.WriteFile("temp/css.liquid", css, 0644)
		if err != nil {
			panic(err)
		}

		js := []byte(`<script>
		alert("xin chao")
	</script>`)
		err = ioutil.WriteFile("temp/js.liquid", js, 0644)
		if err != nil {
			panic(err)
		}

		// Zip ..
		files := []string{"temp/html.liquid", "temp/css.liquid", "temp/js.liquid"}
		output := "temp/theme.zip"

		if err := ZipFiles(output, files); err != nil {
			panic(err)
		}
		fmt.Println("Zipped File:", output)

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
