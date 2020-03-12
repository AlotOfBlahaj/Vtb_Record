package worker

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/plugins/structUtils"
	"github.com/fzxiao233/Vtb_Record/utils"
	"log"
	"strconv"
	"time"
)

func callJsApi(roomId string, status string, filename string) error {
	_, err := utils.HttpGet("http://127.0.0.1:"+utils.Config.ExpressPort+"/api/live?roomId="+roomId+"&status="+status+"&filename="+filename, map[string]string{})
	if err != nil {
		err = fmt.Errorf("call danmaku error %v", err)
		log.Print(err)
		return err
	}
	log.Printf("[Danmaku]%s: %s", roomId, status)
	return nil
}

func getRoomId(targetId string) string {
	var resp []byte
	var err error = nil
	for {
		resp, err = utils.HttpGet("https://api.live.bilibili.com/room/v1/Room/getRoomInfoOld?mid="+targetId, map[string]string{})
		if err != nil {
			log.Printf("cannot get roomid %v", err)
			continue
		}
		respJson, err := simplejson.NewJson(resp)
		if err != nil {
			log.Printf("%s parse json error", targetId)
		}
		if respJson != nil {
			data := respJson.Get("data")
			roomId := strconv.Itoa(data.Get("roomid").MustInt())
			return roomId
		}
	}
}

func StartRecord(video *structUtils.VideoInfo) string {
	if video.UsersConfig.TransBiliId == "" {
		return ""
	}
	filename := video.UsersConfig.TransBiliId + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".txt"
	aFilePath := utils.Config.DownloadDir + "/" + filename
	err := callJsApi(getRoomId(video.UsersConfig.TransBiliId), "1", filename)
	if err != nil {
		return ""
	}
	return aFilePath
}

func CloseRecord(video *structUtils.VideoInfo) {
	if video.UsersConfig.TransBiliId == "" {
		return
	}
	err := callJsApi(getRoomId(video.UsersConfig.TransBiliId), "0", "")
	if err != nil {
		return
	}
}
