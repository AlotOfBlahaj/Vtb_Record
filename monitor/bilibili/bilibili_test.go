package bilibili

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"strings"
	"testing"
)

var FULL_DAY_LIVEROOM = 19193

func TestMain(m *testing.M) {
	config.PrepareConfig()
	m.Run()
}

func TestBilibiliPoller(t *testing.T) {
	poller := &BilibiliPoller{}
	err := poller.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned an error: %s", err)
	}
	liveInfo := poller.IsLiving(FULL_DAY_LIVEROOM)
	if liveInfo == nil {
		t.Fatalf("Failed to detect full-day live room")
	}
	if !strings.Contains(liveInfo.StreamingLink, "bilibili.com") && !strings.HasPrefix(liveInfo.StreamingLink, "http") {
		t.Fatalf("Got malformed streamlink: %s", liveInfo.StreamingLink)
	}
	t.Logf("Successfully got fullday live info: %v", liveInfo)
}
