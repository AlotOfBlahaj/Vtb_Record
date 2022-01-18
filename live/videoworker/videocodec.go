package videoworker

import (
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

func (p *ProcessVideo) convertToMp4(dirpath string) string {
	//livetime := p.liveStartTime.Format("2006-01-02 15:04:05")
	//title := fmt.Sprintf("【%s】", livetime) + p.LiveStatus.Video.Title
	title := p.getFullTitle()
	var videoName string
	if len(p.videoPathList) == 0 {
		log.Warnf("videoPathList is empty!!!! full info: %v", p)
	} else if len(p.videoPathList) > 1 {
		mergedName := utils.ChangeName(title + "_merged.mp4")
		mergedPath := dirpath + "/" + mergedName
		videoName = p.mergeVideo(mergedPath)
	} else {
		mp4Name := utils.ChangeName(title + ".mp4")
		mp4Path := dirpath + "/" + mp4Name
		videoName = p.ts2mp4(p.videoPathList[0], mp4Path)
	}
	return videoName
}

func (p *ProcessVideo) mergeVideo(outpath string) string {
	l := p.videoPathList
	co := "concat:"
	for _, aPath := range l {
		co += aPath + "|"
	}
	logger := log.WithField("video", p.LiveStatus.Video)
	utils.ExecShell("ffmpeg", "-i", co, "-c", "copy", "-f", "mp4", outpath)
	if !utils.IsFileExist(outpath) {
		logger.Warnf("%s the video file don't exist", outpath)
		return ""
	}
	for _, aPath := range l {
		_ = os.Remove(aPath)
	}
	return outpath
}

func (p *ProcessVideo) ts2mp4(tsPath string, outpath string) string {
	logger := log.WithField("video", p.LiveStatus.Video)
	utils.ExecShell("ffmpeg", "-i", tsPath, "-c", "copy", "-f", "mp4", outpath)
	if !utils.IsFileExist(outpath) {
		logger.Warnf("%s the video file don't exist", outpath)
		return ""
	}
	_ = os.Remove(tsPath)
	return outpath
}
