package utils

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

func ExecShell(name string, arg ...string) (string, string) {
	return ExecShellEx(log.NewEntry(log.StandardLogger()), true, name, arg...)
}

func ExecShellEx(entry *log.Entry, redirect bool, name string, arg ...string) (string, string) {
	var stdoutBuf, stderrBuf bytes.Buffer
	co := exec.Command(name, arg...)
	stdoutIn, _ := co.StdoutPipe()
	stderrIn, _ := co.StderrPipe()
	stdout := &stdoutBuf
	stderr := &stderrBuf

	//stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	//stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	_ = co.Start()
	if redirect {
		go func() {
			//_, errStdout = io.Copy(stdout, stdoutIn)
			in := bufio.NewScanner(stdoutIn)
			for in.Scan() {
				stdout.Write(in.Bytes())
				entry.Info(in.Text()) // write each line to your log, or anything you need
			}
		}()
		go func() {
			//_, errStderr = io.Copy(stderr, stderrIn)
			in := bufio.NewScanner(stderrIn)
			for in.Scan() {
				stderr.Write(in.Bytes())
				entry.Info(in.Text()) // write each line to your log, or anything you need
			}
		}()
	} else {
		var errStdout, errStderr error
		go func() {
			_, errStdout = io.Copy(stdout, stdoutIn)
		}()
		go func() {
			_, errStderr = io.Copy(stderr, stderrIn)
		}()
		if errStderr != nil {
			entry.Warnf("%v", errStderr)
		}
		if errStdout != nil {
			entry.Warnf("%v", errStdout)
		}
	}

	_ = co.Wait()
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	//println(outStr + errStr)
	return outStr, errStr
}
