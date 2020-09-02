package monitor

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPoller(t *testing.T) {
	f, _ := os.Open("test.html")
	buf := bytes.NewBuffer(make([]byte, 0))
	io.Copy(buf, f)
	poll := &YoutubePoller{}
	err := poll.parseLiveStatus(buf.String())
	if err != nil {
		t.Errorf("parseLiveStatus() err: %v", err)
	}
	/*t.Run("YoutubePoller", func(t *testing.T) {

	})*/
}
