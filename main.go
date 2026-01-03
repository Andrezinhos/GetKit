package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func Dowloader(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

type Software struct {
	Name      string `json:"name"`
	Repo      string `json:"repo"`
	AssetName string `json:"asset_name"`
	Fallback  string `json:"fallback_url"`
}

func fetchFromRepo(repo, assetName string) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, _ := http.NewRequest("GET", apiURL, nil)
	//req.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Set("User-Agent", "GoClient")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		fmt.Println("Error to decode Json", err)
		return
	}

	home, _ := os.UserHomeDir()
	downloadDir := home + string(os.PathSeparator) + "Downloads" + string(os.PathSeparator)

	fmt.Println("Last Version: ", release.TagName)
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			filePath := downloadDir + asset.Name
			fmt.Println("Downloading From: ", asset.URL)
			if err := Dowloader(asset.URL, filePath); err != nil {
				fmt.Println("Error to Dowload:", err)
			} else {
				fmt.Println("Dowload Complete", asset.Name)
			}
			break
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Use: help to see commands")
		return
	}

	cmd := strings.ToLower(os.Args[1])
	osFlag := strings.TrimPrefix(os.Args[2], "-")
	category := strings.TrimPrefix(os.Args[3], "-")
	requested := strings.ToLower(os.Args[4])

	if cmd == "-v" {
		fmt.Println("GetKit Package Manager v0.3.0 - SQC Tech")
		return
	}
	if cmd != "get" {
		fmt.Println("Unknow Command: ", cmd)
		return
	}

	osMap := map[string]string{
		"win":   "windows",
		"linux": "linux",
		"mac":   "macos",
	}

	osFolder, ok := osMap[osFlag]
	if !ok {
		fmt.Println("Invalid System")
		return
	}

	jsonPath := filepath.Join("C:\\Freighter\\packages", osFolder, category+".json")
	fmt.Println("Loading catalog:", jsonPath)

	file, err := os.Open(jsonPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, _ := io.ReadAll(file)

	var softwares []Software
	if err := json.Unmarshal(data, &softwares); err != nil {
		panic(err)
	}

	home, _ := os.UserHomeDir()
	downloadDir := home + string(os.PathSeparator) + "Downloads" + string(os.PathSeparator)

	found := false
	for _, s := range softwares {
		if strings.ToLower(s.Name) == requested {
			found = true
			fmt.Println("Preparing to Download:", s.Name)
			if s.Repo != "" {
				fetchFromRepo(s.Repo, s.AssetName)
			} else if s.Fallback != "" {
				filePath := downloadDir + s.Name + ".exe"
				fmt.Println("Downloading From:", s.Fallback)
				if err := Dowloader(s.Fallback, filePath); err != nil {
					fmt.Println("Error to Download:", err)
				} else {
					fmt.Println("Download Complete:", s.Name)
				}
			}
			break
		}
	}

	if !found {
		fmt.Println("software not included on catalog:", requested)
	}
}
