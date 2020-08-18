package monitor

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestPoller(t *testing.T) {
	f, _ := os.Open("test.html")
	buf := bytes.NewBuffer(make([]byte, 0))
	io.Copy(buf, f)
	poll := &YoutubePoller{}
	_, err := poll.parseSubscStatus(buf.String())
	if err != nil {
		t.Errorf("parseSubscStatus() err: %v", err)
	}
	/*t.Run("YoutubePoller", func(t *testing.T) {

	})*/
}

func TestPollerBase(t *testing.T) {
	f, _ := os.Open("test_base.html")
	buf := bytes.NewBuffer(make([]byte, 0))
	io.Copy(buf, f)
	poll := &YoutubePoller{}
	_, err := poll.parseBaseStatus(buf.String())
	if err != nil {
		t.Errorf("parseBaseStatus() err: %v", err)
	}
	/*t.Run("YoutubePoller", func(t *testing.T) {

	})*/
}

func TestGetLive(t *testing.T) {
	ctx := &MonitorCtx{
		Client:         http.DefaultClient,
		ExtraModConfig: map[string]interface{}{},
	}
	_, err := getVideoInfo(ctx, "https://muddy-forest-b1aa.vtbrecorder7.workers.dev", "UCb5JxV6vKlYVknoJB8TnyYg")
	if err != nil {
		t.Errorf("parseBaseStatus() err: %v", err)
	}
	/*t.Run("YoutubePoller", func(t *testing.T) {

	})*/
}
