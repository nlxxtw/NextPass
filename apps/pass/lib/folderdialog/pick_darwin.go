//go:build darwin

package folderdialog

import (
	"errors"
	"os/exec"
	"strings"
	"time"
)

func runOSA(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	if err != nil {
		low := strings.ToLower(s + " " + err.Error())
		if strings.Contains(low, "user canceled") || strings.Contains(low, "user cancelled") {
			return "", nil
		}
		if s != "" {
			return "", errors.New(s)
		}
		return "", err
	}
	// 文件夹 POSIX path 常带尾 /
	return strings.TrimRight(s, "/"), nil
}

func escAS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// Pick 弹出 macOS 文件夹选择框
func Pick(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件夹"
	}
	script := `POSIX path of (choose folder with prompt "` + escAS(title) + `")`
	if startDir != "" {
		script = `POSIX path of (choose folder with prompt "` + escAS(title) + `" default location (POSIX file "` + escAS(startDir) + `"))`
	}
	return runOSA(script)
}

// PickFile 弹出 macOS 文件选择框
func PickFile(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件"
	}
	script := `POSIX path of (choose file with prompt "` + escAS(title) + `")`
	if startDir != "" {
		script = `POSIX path of (choose file with prompt "` + escAS(title) + `" default location (POSIX file "` + escAS(startDir) + `"))`
	}
	return runOSA(script)
}

// PickFileOrFolder 先问选文件还是文件夹
func PickFileOrFolder(startDir string) (string, error) {
	choice, err := runOSA(`button returned of (display dialog "选「文件」或「文件夹」发给学生" buttons {"取消", "文件夹", "文件"} default button "文件" with title "NextPass")`)
	if err != nil {
		return "", err
	}
	if choice == "" || choice == "取消" {
		return "", nil
	}
	if choice == "文件夹" {
		return Pick(startDir, "选择要发给学生的文件夹")
	}
	return PickFile(startDir, "选择要发给学生的文件")
}

func CaptureDrop(timeout time.Duration, rect DropRect) (string, error) {
	return "", errors.New("请直接把文件拖到输入框，或点「选文件/选文件夹」")
}

func CancelCaptureDrop() {}

func ClearCaptureDropFlags() {}
