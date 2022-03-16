package native

import (
	"fmt"
	"net/http"
)

type FlvParser struct {
	client *http.Client
}

func (f *FlvParser) ParseStream(url string, filename string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "Chrome/90")
	resp, err := f.client.Do(req)
	defer resp.Body.Close()

	return fmt.Errorf("not implement")
}

func NewFlvParser() *FlvParser {
	return &FlvParser{
		client: &http.Client{},
	}
}
