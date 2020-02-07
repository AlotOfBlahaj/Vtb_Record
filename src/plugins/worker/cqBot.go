package worker

import (
	"Vtb_Record/src/plugins/structUtils"
	"bytes"
	"encoding/json"
	"errors"
	"log"
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
	resp, err := client.Do(req)
	if err != nil {
		log.Print("CQbot error")
	}
	log.Print(resp.StatusCode)
}
func (c *CQMsg) CreateCQMsg(groupId int) {
	c.GroupId = groupId
}
func needCQBot(video *structUtils.VideoInfo) error {
	if !video.UsersConfig.NeedCQBot {
		return errors.New(video.UsersConfig.Name + "needn't download")
	}
	return nil
}
func CQBot(video *structUtils.VideoInfo) error {
	if err := needCQBot(video); err != nil {
		return err
	}
	c := &CQMsg{Message: video.CQBotMsg}
	cc := &CQConfig{
		CQHost:  video.UsersConfig.CQHost,
		CQToken: video.UsersConfig.CQToken,
	}
	for _, GroupId := range video.UsersConfig.QQGroupID {
		c.CreateCQMsg(GroupId)
		cc.sendGroupMsg(c)
		log.Printf("%s|%s send notice to %d", video.Provider, video.UsersConfig.Name, GroupId)
	}
	return nil
}
