package goshopify

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

// RemoveIndexFromSlice ..
func RemoveIndexFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func appendFiles(filename string, zipw *zip.Writer) error {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer os.Chdir(currentDir)

	dir, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("Failed to open %s: %s", filename, err)
	}

	arr := strings.Split(dir, "/")
	filename = arr[len(arr)-1]
	arr = RemoveIndexFromSlice(arr, len(arr)-1)
	dir = strings.Join(arr, "/")
	os.Chdir(dir)

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := zipw.Create(filename)
	if err != nil {
		msg := "Failed to create entry for %s in zip file: %s"
		return fmt.Errorf(msg, filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Failed to write %s to zip: %s", filename, err)
	}
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
