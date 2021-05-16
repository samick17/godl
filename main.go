package main

import (
	"flag"
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
	"path/filepath"
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
/* Get the full version name by user-prompt */
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
/* Get the full version name by prefix */
func getVersionByPrefix(driverList *ResponseDrivers, verPrefix string) string {
	for i := len(driverList.Names) - 1; i >= 0; i-- {
		name := driverList.Names[i]
		version := name[:len(name) - 1]
		if version != "icons" && (strings.HasPrefix(version, verPrefix) || verPrefix == "latest") {
			return version
		}
	}
	return ""
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
/* Get the full version link by platform */
func getDownloadLinkByPlatform(downloadLinks *ResponseDriverDownloadLinks, platform string) string {
	for i := len(downloadLinks.Names) - 1; i >= 0; i-- {
		name := downloadLinks.Names[i]
		version := name
		fileNameWithExt := filepath.Base(version)
		fileName := fileNameWithExt[:strings.Index(fileNameWithExt, ".")]
		if strings.HasSuffix(fileName, platform) {
			return name
		}
	}
	return ""
}

func launchWithPrompt() {
	httpClient = createHttpClient()
	driverList := getDriverVersionList()
	matchedVersion := promptVersions(driverList)
	fmt.Println(matchedVersion)
	downloadLinks := getDriverDownloadLinks(matchedVersion)
	downloadRelPath := promptDownload(downloadLinks)
	fmt.Println(downloadRelPath)
	downloadDriverFromPath(downloadRelPath)
}

func launchWithArgs(version, platform string) {
	httpClient = createHttpClient()
	driverList := getDriverVersionList()
	matchedVersion := getVersionByPrefix(driverList, version)
	fmt.Println(matchedVersion)
	downloadLinks := getDriverDownloadLinks(matchedVersion)
	downloadRelPath := getDownloadLinkByPlatform(downloadLinks, platform)
	fmt.Println(downloadRelPath)
	downloadDriverFromPath(downloadRelPath)
}

func printAuthor() {
	colorReset := "\033[0m"

    colorRed := "\033[31m"
    // colorGreen := "\033[32m"
    // colorYellow := "\033[33m"
    // colorBlue := "\033[34m"
    // colorPurple := "\033[35m"
    // colorCyan := "\033[36m"
    // colorWhite := "\033[37m"
	fmt.Println("+-----------------------------+")
	fmt.Println("    ChromeDriver downloader")
	fmt.Println("    Author: ", string(colorRed), "Samick", string(colorReset))
	fmt.Println("    Github: ", string(colorRed), "https://github.com/samick17", string(colorReset))
	fmt.Println("+-----------------------------+")
}

func main() {
	printAuthor()
	fmt.Println("Please choose the version to download:")
	var version string
    flag.StringVar(&version, "version", "", "The version")
    var platform string
    flag.StringVar(&platform, "platform", "", "The platform (linux64, mac64, win32)")
	flag.Parse()
	if version == "" || platform == "" {
		launchWithPrompt()
	} else {
		launchWithArgs(version, platform)
	}
}
