package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/reujab/wallpaper"
)

type ImageList struct {
	Images []Image
}

type Image struct {
	Urlbase string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Environment File couldn't be loaded")
	}

	bingurl := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"

	resp, err := http.Get(bingurl)
	if err != nil {
		log.Fatal("Bing not reachable")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Bing Response empty")
	}

	var iml ImageList
	err = json.Unmarshal(body, &iml)
	if err != nil {
		log.Fatal("JSON Parsing error")
	}

	if len(iml.Images) < 1 {
		log.Fatal("Image not found")
	}

	u, err := url.Parse(iml.Images[0].Urlbase)
	if err != nil {
		log.Fatal("Url parse error")
	}

	baseimg := u.Query().Get("id")
	postfix := "1920x1080"

	if os.Getenv("WALLPAPER_UHD") == "yes" {
		postfix = "UHD"
	}

	image := fmt.Sprintf("%s_%s.jpg", baseimg, postfix)
	downloadurl := fmt.Sprintf("http://www.bing.com%s_%s.jpg", iml.Images[0].Urlbase, postfix)

	bingWallperPath := os.Getenv("PICTURE_DOWNLOAD_FOLDER")
	targetPath := filepath.Join(bingWallperPath, image)

	//Ensure Directory exists
	if err = os.MkdirAll(bingWallperPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	//Check if file doesn't exist
	if fileExists(targetPath) {
		fmt.Printf("Image %s", image)
		fmt.Println(" has already been downloaded")
	} else {
		err = downloadWallpaper(downloadurl, targetPath)
		if err != nil {
			log.Fatal(err)
		}
		wallpaper.SetFromFile(targetPath)
	}
	fmt.Println(downloadurl)
	fmt.Println(baseimg)

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func downloadWallpaper(url string, target string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
