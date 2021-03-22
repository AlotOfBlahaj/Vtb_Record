package utils

import (
	"context"
	_ "github.com/rclone/rclone/backend/all"
	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/sync"
)

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
