package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/haicaodac/goshopify"
)

var app goshopify.AppShopify

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func removeIndexFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func appendFiles(source string, zipw *zip.Writer) error {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer os.Chdir(currentDir)

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	dir, err := filepath.Abs(source)
	if err != nil {
		log.Fatalf("Failed to open %s: %s", source, err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	} else {
		arr := strings.Split(dir, "/")
		log.Println(arr)
		arr = removeIndexFromSlice(arr, len(arr)-1)
		dir = strings.Join(arr, "/")
	}

	os.Chdir(dir)

	filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		log.Println(os.Getwd())
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			log.Println(baseDir)
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}
		log.Println(header.Name)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	// file, err := os.Open(source)
	// if err != nil {
	// 	return fmt.Errorf("Failed to open %s: %s", source, err)
	// }
	// defer file.Close()

	// wr, err := zipw.Create(source)
	// if err != nil {
	// 	msg := "Failed to create entry for %s in zip file: %s"
	// 	return fmt.Errorf(msg, source, err)
	// }

	// if _, err := io.Copy(wr, file); err != nil {
	// 	return fmt.Errorf("Failed to write %s to zip: %s", source, err)
	// }
	return nil
}

// ZipFiles ..
func ZipFiles(files []string, zipName string) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(zipName, flags, 0644)
	if err != nil {
		log.Fatalf("Failed to open zip for writing: %s", err)
		return err
	}
	defer file.Close()

	zipw := zip.NewWriter(file)
	defer zipw.Close()

	for _, filename := range files {
		if err := appendFiles(filename, zipw); err != nil {
			log.Fatalf("Failed to add file %s to zip: %s", filename, err)
			return err
		}

		// Remove file after append to zip
		// if err := os.Remove(filename); err != nil {
		// 	log.Fatalf("Failed to remove file %s: %s", filename, err)
		// }
	}

	// Move zip to temp
	// if err := os.Rename(zipName, "temp/"+zipName); err != nil {
	// 	log.Fatalf("Failed to move zip file %s: %s", zipName, err)
	// }

	return nil

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
			os.Mkdir("temp", 0644)
		}

		html := []byte(`<html>
		<head>
			<title>Hi Nora</title>
		</head>
		<body>
			<h1>Minh</h1>
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
		alert("Hello Nora")
	</script>`)
		err = ioutil.WriteFile("temp/js.liquid", js, 0644)
		if err != nil {
			panic(err)
		}
		zipName := "temp/Narrative.zip"
		files := []string{"temp/theme2/"}
		if err := ZipFiles(files, zipName); err != nil {
			log.Println("Failed to zip files")
		}

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
