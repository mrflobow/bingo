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

	"github.com/reujab/wallpaper"
	"golang.org/x/sys/windows/registry"
)

type ImageList struct {
	Images []Image
}

type Image struct {
	Urlbase string
}

func main() {
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
	image := fmt.Sprintf("%s_1920x1080.jpg", baseimg)
	downloadurl := fmt.Sprintf("http://www.bing.com%s_1920x1080.jpg", iml.Images[0].Urlbase)

	picturesFolder, err := getPicturesFolder()
	if err != nil {
		log.Fatal(err)
	}

	bingWallperPath := filepath.Join(picturesFolder, "BingWallpaper")
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

func getPicturesFolder() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("My Pictures")

	return s, err
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
