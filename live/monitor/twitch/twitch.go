package twitch

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/monitor/base"
)

type Twitch struct {
	base.BaseMonitor
	APIUrl string
}

func (t Twitch) getLiveStatus() error {
	_, err := t.Ctx.HttpGet(t.APIUrl, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}

func (t Twitch) CheckLive(userConfig config.UsersConfig) {
	t.APIUrl = "https://api.twitch.tv/helix/streams?user_login=" + userConfig.TargetId
}
