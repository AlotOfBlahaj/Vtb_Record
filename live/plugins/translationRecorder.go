package plugins

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/fzxiao233/Vtb_Record/live/videoworker"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func callJsAPI(roomID string, status string, filename string) error {
	_, err := utils.HttpGet(nil, "http://127.0.0.1:"+utils.Config.ExpressPort+"/api/live?roomID="+roomID+"&status="+status+"&filename="+filename, map[string]string{})
	if err != nil {
		err = fmt.Errorf("call danmaku error %v", err)
		log.Warn(err)
		return err
	}
	log.Debugf("[Danmaku]%s: %s", roomID, status)
	return nil
}

func getRoomId(targetId string) string {
	var resp []byte
	var err error = nil
	for {
		resp, err = utils.HttpGet(nil, "https://api.live.bilibili.com/room/v1/Room/getRoomInfoOld?mid="+targetId, map[string]string{})
		if err != nil {
			log.Errorf("cannot get roomid %v", err)
			continue
		}
		respJson, err := simplejson.NewJson(resp)
		if err != nil {
			log.Errorf("%s parse json error", targetId)
		}
		if respJson != nil {
			data := respJson.Get("data")
			roomId := strconv.Itoa(data.Get("roomid").MustInt())
			return roomId
		}
	}
}

type PluginTranslationRecorder struct {
}

func (p *PluginTranslationRecorder) LiveStart(process *videoworker.ProcessVideo) error {
	return nil
}

func (p *PluginTranslationRecorder) DownloadStart(process *videoworker.ProcessVideo) error {
	video := process.LiveStatus.Video
	if video.UsersConfig.TransBiliId == "" {
		video.TransRecordPath = ""
		return nil
	}
	filename := video.UsersConfig.TransBiliId + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".txt"
	aFilePath := utils.Config.DownloadDir + "/" + filename
	err := callJsAPI(getRoomId(video.UsersConfig.TransBiliId), "1", filename)
	if err != nil {
		return err
	}
	video.TransRecordPath = aFilePath
	return nil
}

func (p *PluginTranslationRecorder) LiveEnd(process *videoworker.ProcessVideo) error {
	video := process.LiveStatus.Video
	if video.UsersConfig.TransBiliId == "" {
		return nil
	}
	err := callJsAPI(getRoomId(video.UsersConfig.TransBiliId), "0", "")
	if err != nil {
		return err
	}
	return nil
}
