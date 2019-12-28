package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func HttpGet(url string) []byte {
	var client *http.Client
	if Config.EnableProxy == true {
		client = createSOCKS5Proxy()
	} else {
		client = http.DefaultClient
	}
	req, err := http.NewRequest("GET", url, nil)
	CheckError(err, "request html error")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:60.0) Gecko/20100101 Firefox/60.0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	res, err := client.Do(req)
	CheckError(err, "request html error")
	htmlBody, _ := ioutil.ReadAll(res.Body)
	return htmlBody
}
func createSOCKS5Proxy() *http.Client {
	proxyUrl, _ := url.Parse("socks5://" + Config.Proxy)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	//adding the Transport object to the http Client
	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
	return client
}
func IsFileExist(aFilepath string) bool {
	if _, err := os.Stat(aFilepath); err == nil {
		return true
	} else {
		return false
	}
}
func GenerateFilepath(UserName string, VideoTitle string) string {
	pathSlice := []string{Config.DownloadDir, UserName, VideoTitle}
	aFilepath := strings.Join(pathSlice, "/")
	if IsFileExist(aFilepath) {
		return changeName(aFilepath)
	} else {
		return aFilepath
	}
}
func changeName(aFilepath string) string {
	dir, file := filepath.Split(aFilepath)
	ext := path.Ext(file)
	filename := path.Base(file)
	filename += string(time.Now().Unix())
	return dir + filename + ext
}
func GetTimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
