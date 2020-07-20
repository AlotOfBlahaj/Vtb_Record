package utils

import (
	"bytes"
	"fmt"
	"github.com/fzxiao233/Go-Emoji-Utils"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//var client *http.Client

func init() {
	//client = createClient()
}

func MapToStruct(mapVal map[string]interface{}, structVal interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           structVal,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	decoder.Decode(mapVal)
	return nil
}

func HttpGetBuffer(client *http.Client, url string, header map[string]string, buf *bytes.Buffer) (*bytes.Buffer, error) {
	if client == nil {
		client = &http.Client{}
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:60.0) Gecko/20100101 Firefox/60.0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		err = fmt.Errorf("HttpGet error %w", err)
		log.Warn(err)
		return nil, err
	}

	if res.StatusCode != 200 && res.StatusCode != 206 {
		err = fmt.Errorf("HttpGet status code error %d", res.StatusCode)
		//log.Warn(err)
		return nil, err
	}

	//log.Infof("%v", res.ContentLength)

	//var htmlBody []byte
	if res.ContentLength >= 0 {
		if buf == nil {
			buf = bytes.NewBuffer(make([]byte, res.ContentLength))
		}
		buf.Reset()
		if int64(buf.Cap()) < res.ContentLength {
			buf.Grow(int(res.ContentLength) - buf.Cap())
		}
		//buf := bytes.NewBuffer(make([]byte, 0, res.ContentLength))
		n, err := io.Copy(buf, res.Body)
		if err != nil {
			return nil, err
		}
		if n != res.ContentLength {
			log.Warnf("Got unexpected payload: expected: %v, got %v", res.ContentLength, n)
		}
		//htmlBody = buf.Bytes()
	} else {
		if buf == nil {
			buf = bytes.NewBuffer(make([]byte, 2048))
		}
		buf.Reset()
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			return nil, err
		}
		//htmlBody, _ = ioutil.ReadAll(res.Body)
	}
	return buf, nil
}

func HttpGet(client *http.Client, url string, header map[string]string) ([]byte, error) {
	buf, err := HttpGetBuffer(client, url, header, nil)
	if err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func IsFileExist(aFilepath string) bool {
	if _, err := os.Stat(aFilepath); err == nil {
		return true
	} else {
		//log.Errorf("File not exist %s, stat err %s", aFilepath, err)
		return false
	}
}
func GenerateFilepath(DownDir string, VideoTitle string) string {
	pathSlice := []string{DownDir, VideoTitle}
	aFilepath := strings.Join(pathSlice, "/")
	/*if IsFileExist(aFilepath) {
		return ChangeName(aFilepath)
	} else {
		return aFilepath
	}*/
	return ChangeName(aFilepath)
}
func MakeDir(dirPath string) string {
	if !IsFileExist(dirPath) {
		//err := os.MkdirAll(dirPath, 0775)
		err := MkdirAll(dirPath)
		if err != nil {
			log.Fatalf("mkdir error: %s", dirPath)
		}
	}
	return dirPath
}

func AddSuffix(aFilepath string, suffix string) string {
	dir, file := filepath.Split(aFilepath)
	ext := path.Ext(file)
	filename := strings.TrimSuffix(path.Base(file), ext)
	filename += "_"
	filename += suffix
	return dir + filename + ext
}

func ChangeName(aFilepath string) string {
	return AddSuffix(aFilepath, strconv.FormatInt(time.Now().Unix(), 10))
}
func GetTimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
func RemoveIllegalChar(Title string) string {
	illegalChars := []string{"|", "/", "\\", ":", "?"}
	Title = emoji.RemoveAll(Title)
	for _, char := range illegalChars {
		Title = strings.ReplaceAll(Title, char, "#")
	}
	return Title
}

func I2b(i int) bool {
	if i != 0 {
		return true
	} else {
		return false
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RPartition(s string, sep string) (string, string, string) {
	parts := strings.SplitAfter(s, sep)
	if len(parts) == 1 {
		return "", "", parts[0]
	}
	return strings.Join(parts[0:len(parts)-1], ""), sep, parts[len(parts)-1]
}

func RandChooseStr(arr []string) string {
	return arr[rand.Intn(len(arr))]
}
