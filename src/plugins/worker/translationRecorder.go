package worker

import (
	"Vtb_Record/src/plugins/structUtils"
	"Vtb_Record/src/utils"
	"github.com/bitly/go-simplejson"
	"log"
	"strconv"
	"time"
)

func callJsApi(roomId string, status string, filename string) {
	utils.HttpGet("http://127.0.0.1:" + utils.Config.ExpressPort + "/api/live?roomId=" + roomId + "&status=" + status + "&filename=" + filename)
	log.Printf("[Danmaku]%s: %s", roomId, status)
}

func getRoomId(targetId string) string {
	resp := utils.HttpGet("https://api.live.bilibili.com/room/v1/Room/getRoomInfoOld?mid=" + targetId)
	respJson, err := simplejson.NewJson(resp)
	if err != nil {
		log.Printf("%s parse json error", targetId)
	}
	data := respJson.Get("data")
	roomId := strconv.Itoa(data.Get("roomid").MustInt())
	return roomId
}

func StartRecord(video *structUtils.VideoInfo) string {
	if video.UsersConfig.TransBiliId == "" {
		return ""
	}
	filename := video.UsersConfig.TransBiliId + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".txt"
	aFilePath := utils.GenerateDownloadDir(video.UsersConfig.Name) + "/" + filename
	go callJsApi(getRoomId(video.UsersConfig.TransBiliId), "1", filename)
	return aFilePath
}

func CloseRecord(video *structUtils.VideoInfo) {
	if video.UsersConfig.TransBiliId == "" {
		return
	}
	callJsApi(getRoomId(video.UsersConfig.TransBiliId), "0", "")
}
