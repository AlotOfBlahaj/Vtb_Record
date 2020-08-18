package downloader

import (
	"bufio"
	"bytes"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/fzxiao233/Vtb_Record/live/interfaces"
	"github.com/fzxiao233/Vtb_Record/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

func addStreamlinkProxy(co []string, proxy string) []string {
	co = append(co, "--http-proxy", "socks5://"+proxy)
	return co
}

type DownloaderStreamlink struct {
	Downloader
}

func (d *DownloaderStreamlink) StartDownload(video *interfaces.VideoInfo, proxy string, cookie string, filepath string) error {
	_arg, ok := video.UsersConfig.ExtraConfig["StreamLinkArgs"]
	arg := []string{}
	if ok {
		for _, a := range _arg.([]interface{}) {
			arg = append(arg, a.(string))
		}
	}
	//arg = append(arg, []string{"--force", "-o", filepath}...)
	if proxy != "" {
		arg = addStreamlinkProxy(arg, proxy)
	}
	arg = append(arg, video.Target, config.Config.DownloadQuality)
	logger := log.WithField("video", video)
	logger.Infof("start to download %s, command %s", filepath, arg)
	//utils.ExecShellEx(logger, true, "streamlink", arg...)
	downloader := &StreamlinkDownload{
		Logger:            logger,
		Video:             video,
		Filepath:          filepath,
		StreamlinkCommand: arg,
	}
	return downloader.doDownload()
}

type StreamlinkDownload struct {
	Logger            *log.Entry
	Video             *interfaces.VideoInfo
	Filepath          string
	StreamlinkCommand []string
}

func (d *StreamlinkDownload) doDownload() error {
	out := utils.GetWriter(d.Filepath)
	defer out.Close()
	d.StreamlinkCommand = append(d.StreamlinkCommand, []string{"--force", "--stdout"}...)
	var stderrBuf bytes.Buffer
	co := exec.Command("streamlink_", d.StreamlinkCommand...)
	stdoutIn, _ := co.StdoutPipe()
	stderrIn, _ := co.StderrPipe()
	stderr := &stderrBuf

	_ = co.Start()
	go func() {
		//_, errStderr = io.Copy(stderr, stderrIn)
		in := bufio.NewScanner(stderrIn)
		for in.Scan() {
			stderr.Write(in.Bytes())
			d.Logger.Info(in.Text()) // write each line to your log, or anything you need
		}
	}()
	errChan := make(chan error)
	go func() {
		_, errStdout := io.Copy(out, stdoutIn)
		if errStdout != nil {
			d.Logger.WithError(errStdout).Warn("Error during writing streamlink video")
		}
		errChan <- errStdout
	}()

	_ = co.Wait()
	err := <-errChan
	return err
}
