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
	_ = co.Start()
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	if errStderr != nil {
		log.Printf("%v", errStderr)
	}
	if errStdout != nil {
		log.Printf("%v", errStdout)
	}
	_ = co.Wait()
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	//println(outStr + errStr)
	return outStr + errStr
}
