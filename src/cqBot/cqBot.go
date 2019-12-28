package cqBot

import (
	"Vtb_Record/src/utils"
	"bytes"
	"encoding/json"
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

// Todo: cqBot support
func (c CQConfig) sendGroupMsg(msg CQMsg) {
	JsonMsg, _ := json.Marshal(msg)
	resp, err := http.Post("http://"+c.CQHost+"/send_group_msg",
		"application/json;charset=utf-8",
		bytes.NewBuffer(JsonMsg))
	utils.CheckError(err, "CQBot")
	println(resp)
}
func CreateCQMsg() {

}
func CQBot(video utils.VideoInfo) {

}
