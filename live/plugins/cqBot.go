package plugins

import (
	"bytes"
	"encoding/json"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type CQConfig struct {
	CQHost  string
	CQToken string
}
type CQMsg struct {
	GroupId int    `json:"group_id"`
	Message string `json:"message"`
}

func (cc *CQConfig) sendGroupMsg(msg *CQMsg) {
	client := &http.Client{}
	JsonMsg, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", "http://"+cc.CQHost+"/send_group_msg", bytes.NewBuffer(JsonMsg))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cc.CQToken)
	_, err := client.Do(req)
	if err != nil {
		log.Warnf("CQbot error")
	} else {
		log.Infof("%s", msg.Message)
	}
}

type PluginCQBot struct {
	sentMsg map[string]map[int]int
}

func CreateLiveMsg(v *interfaces.VideoInfo) string {
	return "[直播提示]" + "[" + v.Provider + "]" + v.Title + "正在直播" + "链接: " + v.Target + " [CQ:at,qq=all]"
}

type CQJsonConfig struct {
	NeedCQBot bool
	QQGroupID []int
	CQHost    string
	CQToken   string
}

func (p *PluginCQBot) LiveStart(process *videoworker.ProcessVideo) error {
	if p.sentMsg == nil {
		p.sentMsg = make(map[string]map[int]int)
	}
	video := process.LiveStatus.Video
	config := CQJsonConfig{}
	cqConfig, ok := video.UsersConfig.ExtraConfig["CQConfig"]
	if !ok {
		return nil
	}
	err := utils.MapToStruct(cqConfig.(map[string]interface{}), &config)
	if err != nil {
		return err
	}

	if !config.NeedCQBot {
		log.Tracef(video.UsersConfig.Name + " needn't cq")
		return nil
	}

	msg := CreateLiveMsg(video)
	c := &CQMsg{Message: msg}
	cc := &CQConfig{
		CQHost:  config.CQHost,
		CQToken: config.CQToken,
	}
	for _, GroupId := range config.QQGroupID {
		sentGroupIds := p.sentMsg[msg]
		if sentGroupIds == nil {
			p.sentMsg[msg] = make(map[int]int)
		}
		_, ok = sentGroupIds[GroupId]
		if ok {
			log.Infof("%s|%s cancel to send msg: %s", video.Provider, video.UsersConfig.Name, msg)
			continue
		}
		c.GroupId = GroupId
		cc.sendGroupMsg(c)
		log.Infof("%s|%s send notice to %d", video.Provider, video.UsersConfig.Name, GroupId)

		p.sentMsg[msg][GroupId] = 1
	}
	return nil
}

func (p *PluginCQBot) DownloadStart(process *videoworker.ProcessVideo) error {
	return nil
}

func (p *PluginCQBot) LiveEnd(process *videoworker.ProcessVideo) error {
	return nil
}
