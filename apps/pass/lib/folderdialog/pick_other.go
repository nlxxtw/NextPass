//go:build !windows && !darwin

package folderdialog

import (
	"errors"
	"time"
)

func Pick(startDir, title string) (string, error) {
	return "", errors.New("当前系统请在输入框粘贴目录绝对路径，例如 /home/user/docs")
}

func PickFile(startDir, title string) (string, error) {
	return "", errors.New("当前系统请在输入框粘贴文件绝对路径")
}

func PickFileOrFolder(startDir string) (string, error) {
	return "", errors.New("当前系统请在输入框粘贴文件或文件夹绝对路径")
}

func CaptureDrop(timeout time.Duration, rect DropRect) (string, error) {
	return "", errors.New("当前系统不支持拖放选路径，请粘贴绝对路径")
}

func CancelCaptureDrop() {}

func ClearCaptureDropFlags() {}
