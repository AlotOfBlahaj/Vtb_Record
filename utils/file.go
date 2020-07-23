package utils

import (
	"context"
	_ "github.com/rclone/rclone/backend/all"
	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/sync"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type FileType int

const (
	FILE_NORMAL FileType = 0
	FILE_RCLONE FileType = 1
)

type FsType int

const (
	FS_OS     FsType = 0
	FS_RCLONE FsType = 1
)

type FileSystem struct {
	FsType FsType
	FsData interface{}
}

type FsRcloneData struct {
	rcloneFs *fs.Fs
}

func MkdirAll(path string) error {
	fdst := cmd.NewFsDir([]string{path})
	err := operations.Mkdir(context.Background(), fdst, "")
	return err
}

func MoveFiles(src string, dst string) error {
	fsrc, srcFileName, fdst := cmd.NewFsSrcFileDst([]string{src, dst})
	if srcFileName == "" {
		return sync.MoveDir(context.Background(), fdst, fsrc, false, false)
	}
	return operations.MoveFile(context.Background(), fdst, fsrc, srcFileName, srcFileName)
}

func GetWriter(dst string) io.WriteCloser {
	reader, writer := io.Pipe()
	fdst, dstFileName := cmd.NewFsDstFile([]string{dst})
	go func() {
		_, err := operations.Rcat(context.Background(), fdst, dstFileName, reader, time.Now())
		if err != nil {
			log.Warnf("Rcat [%s] Error! err: %s", dst, err)
		}
	}()
	return writer
}
