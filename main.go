package main

import (
	"fmt"
	"net/http"
	"time"
	"io/ioutil"
	"encoding/xml"
	"os"
	"io"
	"bufio"
	"strings"
	"path"
)

type Client struct {
	Trans *http.Transport
}

func createHttpClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	return client
}

var httpClient *http.Client

func getChromeDriversUrl() string {
	return "http://chromedriver.storage.googleapis.com/?delimiter=/&prefix="
}
func getDriverUrlByVersion(version string) string {
	return fmt.Sprintf("http://chromedriver.storage.googleapis.com/?delimiter=/&prefix=%s/", version)
}
type ResponseDrivers struct {
	// Result xml.Name `xml:"ListBucketResult"`
	Names []string `xml:"CommonPrefixes>Prefix"`
	// Key string `xml:"Contents>Key"`
	// Key string `xml:"Contents>Key"`
	// Name string `xml:"Name"`
	// Prefix string `xml:"Prefix"`
}
type ResponseDriverDownloadLinks struct {
	Names []string `xml:"Contents>Key"`
}

func downloadFile(url, filepath string) error {
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
func mkdirp(filepath string) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
	    os.Mkdir(filepath, os.ModePerm)
	}
}
func getDriverVersionList() *ResponseDrivers {
	url := getChromeDriversUrl()
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// content := string(body)
	// fmt.Println(content)
	// v := make(map[string]interface{})
	data := &ResponseDrivers{}
	err = xml.Unmarshal(body, data)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}
	return data
}
func getDriverDownloadLinks(version string) *ResponseDriverDownloadLinks {
	url := getDriverUrlByVersion(version)
	fmt.Println(url)
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// content := string(body)
	// fmt.Println(content)
	// v := make(map[string]interface{})
	data := &ResponseDriverDownloadLinks{}
	err = xml.Unmarshal(body, data)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}
	return data
}
func promptVersions(driverList *ResponseDrivers) string {
	downloadableVersions := make(map[string]string)
	for index, name := range driverList.Names {
		version := name[:len(name) - 1]
		if version != "icons" {
			downloadableVersions[fmt.Sprintf("%d", index+1)] = version
			fmt.Println(fmt.Sprintf("%d: %s", index+1, version))
		}
	}
	/* Prompt to download */
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("-> ")
    text, _ := reader.ReadString('\n')
    text = strings.Replace(text, "\n", "", -1)
    fmt.Println(fmt.Sprintf("Your input: %s", text))
    if matchedVersion, ok := downloadableVersions[text]; ok {
    	return matchedVersion
    } else {
    	return promptVersions(driverList)
    }
}
func downloadDriverFromPath(relPath string) {
	url := fmt.Sprintf("%s%s", "http://chromedriver.storage.googleapis.com/", relPath)
	fileName := relPath[strings.LastIndex(relPath, "/")+1:]
	cwd, err := os.Getwd()
	if err != nil {
	    fmt.Println(err)
	    return
	}
	destPath := path.Join(cwd, "downloads")
	mkdirp(destPath)
	downloadPath := path.Join(destPath, fileName)
	downloadFile(url, downloadPath)
}
func promptDownload(downloadLinks *ResponseDriverDownloadLinks) string {
	fmt.Println(downloadLinks)
	downloadableVersions := make(map[string]string)
	for index, name := range downloadLinks.Names {
		version := name
		downloadableVersions[fmt.Sprintf("%d", index+1)] = version
		fmt.Println(fmt.Sprintf("%d: %s", index+1, version))
	}
	/* Prompt to download */
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("-> ")
    text, _ := reader.ReadString('\n')
    text = strings.Replace(text, "\n", "", -1)
    fmt.Println(fmt.Sprintf("Your input: %s", text))
    if matchedVersion, ok := downloadableVersions[text]; ok {
    	return matchedVersion
    } else {
    	return promptDownload(downloadLinks)
    }
}

func main() {
	httpClient = createHttpClient()
	driverList := getDriverVersionList()
	
	matchedVersion := promptVersions(driverList)
	
    fmt.Println(fmt.Sprintf("Try to download: %s", matchedVersion))
    // url := getDriverUrlByVersion(matchedVersion)
    // fmt.Println(fmt.Sprintf("Download from %s to %s", url, destPath))
    downloadLinks := getDriverDownloadLinks(matchedVersion)
    downloadRelPath := promptDownload(downloadLinks)
    fmt.Println(downloadRelPath)
    downloadDriverFromPath(downloadRelPath)
}
