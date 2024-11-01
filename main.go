package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

type VLCStatus struct {
	Information struct {
		Category struct {
			Meta struct {
				Title   string `json:"title"`
				Artist  string `json:"artist"`
				Album   string `json:"album"`
				Artwork string `json:"artwork_url"`
			} `json:"meta"`
		} `json:"category"`
	} `json:"information"`
}

func main() {
	vlcURL := flag.String("url", "http://localhost:8080/requests/status.json", "VLC API URL")
	password := flag.String("pass", "", "Lua HTTP Password")
	flag.Parse()

	var status VLCStatus
	GetVLCStatus(&status, *vlcURL, *password)
	if status.Information.Category.Meta.Title == "" &&
		status.Information.Category.Meta.Artist == "" &&
		status.Information.Category.Meta.Album == "" {
		return
	}

	log.Printf("Title: %s\n", status.Information.Category.Meta.Title)
	log.Printf("Artist: %s\n", status.Information.Category.Meta.Artist)
	log.Printf("Album: %s\n", status.Information.Category.Meta.Album)
	Start(&status)
}

func GetVLCStatus(status *VLCStatus, vlcURL, password string) *VLCStatus {
	client := &http.Client{}
	req, err := http.NewRequest("GET", vlcURL, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.SetBasicAuth("", password)

	resp, err := client.Do(req)
	log.Println(resp.Status)
	if err != nil {
		log.Fatalf("Failed to get VLC status: %v", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("May be missing password: %v\n", resp.Status)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, &status); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}
	return status
}

func Start(status *VLCStatus) {
	var (
		title        = status.Information.Category.Meta.Title
		artist       = status.Information.Category.Meta.Artist
		album        = status.Information.Category.Meta.Album
		artwork_path = status.Information.Category.Meta.Artwork
	)

	contents := url.QueryEscape(title + " - " + artist + "\nAlbum: " + album)
	intent := "https://twitter.com/intent/tweet?text=" + contents

	if runtime.GOOS == "windows" {
		ExecCmd("cmd", "/c", "start", "chrome", intent)
		ExecCmd("cmd", "/c", "start", "chrome", artwork_path)
	} else if runtime.GOOS == "darwin" {
		ExecCmd("open", intent)
		ExecCmd("open", artwork_path)
	} else {
		log.Fatalln(UnSupportedOSError)
	}

}

func ExecCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}

var UnSupportedOSError = errors.New("Your OS is not supported: " + runtime.GOOS)
