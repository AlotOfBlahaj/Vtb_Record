package utils

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
)

func ExecShell(name string, arg ...string) string {
	var stdoutBuf, stderrBuf bytes.Buffer
	co := exec.Command(name, arg...)
	stdoutIn, _ := co.StdoutPipe()
	stderrIn, _ := co.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := co.Start()
	if err != nil {
		log.Fatalf("ffmpeg failed with '%s'\n", err)
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	err = co.Wait()
	if err != nil {
		log.Fatalf("ffmpeg failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	//println(outStr + errStr)
	return outStr + errStr
}
