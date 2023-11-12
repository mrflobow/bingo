package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/reujab/wallpaper"
)

type ImageList struct {
	Images []Image
}

type Image struct {
	Urlbase string
}

type bingoConfig struct {
	pictureDownloadPath string
	uhd                 bool
}

func NewBingoConfig() *bingoConfig {
	c := bingoConfig{"", true}
	return &c
}

func (p *bingoConfig) loadConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cos := runtime.GOOS

	switch cos {
	case "windows":
		p.pictureDownloadPath = path.Join(home, "Pictures", "Bing")
	case "darwin":
		p.pictureDownloadPath = path.Join(home, "Pictures", "Bing")
	default:
	}

	return nil
}

func (p *bingoConfig) createDirectories() error {
	if err := os.MkdirAll(p.pictureDownloadPath, os.ModePerm); err != nil {
		return err
	}
	return nil
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {

	useUHD := flag.Bool("uhd", true, "Use UHD Wallpaper?")
	flag.Parse()

	confPtr := NewBingoConfig()
	if err := confPtr.loadConfig(); err != nil {
		log.Fatal("Error Loading configuration")
	}
	confPtr.uhd = *useUHD

	if err := confPtr.createDirectories(); err != nil {
		log.Fatal("Error creating directories")
	}

	bingurl := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"

	resp, err := http.Get(bingurl)
	if err != nil {
		log.Fatal("Bing not reachable")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
	var postfix string

	if confPtr.uhd {
		postfix = "UHD"
	} else {
		postfix = "1920x1080"
	}

	image := fmt.Sprintf("%s_%s.jpg", baseimg, postfix)
	downloadurl := fmt.Sprintf("http://www.bing.com%s_%s.jpg", iml.Images[0].Urlbase, postfix)

	targetPath := filepath.Join(confPtr.pictureDownloadPath, image)
	fmt.Println(targetPath)
	//Check if file doesn't exist
	if fileExists(targetPath) {
		fmt.Printf("Image %s", image)
		fmt.Println(" has already been downloaded")
	} else {
		err = downloadWallpaper(downloadurl, targetPath)
		if err != nil {
			log.Fatal(err)
		}

	}
	wallpaper.SetFromFile(targetPath)
	fmt.Println("Image set as wallpaper")
}
