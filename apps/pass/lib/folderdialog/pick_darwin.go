//go:build darwin

package folderdialog

import (
	"errors"
	"os/exec"
	"strings"
	"time"
)

func runOSA(script string) (string, error) {
	// 用 stdin 喂多行脚本，比多个 -e 稳；并尽量把对话框提到前台
	full := "try\n" +
		"  tell application \"SystemUIServer\" to activate\n" +
		"end try\n" +
		"delay 0.05\n" +
		script + "\n"
	cmd := exec.Command("osascript")
	cmd.Stdin = strings.NewReader(full)
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
	return strings.TrimRight(s, "/"), nil
}

func escAS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func chooseFolderScript(title, startDir string) string {
	if startDir != "" {
		return `POSIX path of (choose folder with prompt "` + escAS(title) + `" default location (POSIX file "` + escAS(startDir) + `"))`
	}
	return `POSIX path of (choose folder with prompt "` + escAS(title) + `")`
}

func chooseFileScript(title, startDir string) string {
	if startDir != "" {
		return `POSIX path of (choose file with prompt "` + escAS(title) + `" default location (POSIX file "` + escAS(startDir) + `"))`
	}
	return `POSIX path of (choose file with prompt "` + escAS(title) + `")`
}

// Pick 弹出 macOS 文件夹选择框
func Pick(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件夹"
	}
	p, err := runOSA(chooseFolderScript(title, startDir))
	if err != nil && startDir != "" {
		// default location 无效时（权限/路径）降级再弹一次
		return runOSA(chooseFolderScript(title, ""))
	}
	return p, err
}

// PickFile 弹出 macOS 文件选择框
func PickFile(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件"
	}
	p, err := runOSA(chooseFileScript(title, startDir))
	if err != nil && startDir != "" {
		return runOSA(chooseFileScript(title, ""))
	}
	return p, err
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

// FinderSelection 读取访达当前选中项的 POSIX 路径（拖放后常仍保持选中，可自动识别）
func FinderSelection() (string, error) {
	script := `tell application "Finder"
  if (count of (selection as list)) is 0 then
    return ""
  end if
  return POSIX path of (item 1 of (selection as alias list))
end tell`
	return runOSA(script)
}

func CaptureDrop(timeout time.Duration, rect DropRect) (string, error) {
	// 浏览器拿不到绝对路径时：优先用访达当前选中项（用户刚拖的文件通常仍在选中）
	if p, err := FinderSelection(); err == nil && strings.TrimSpace(p) != "" {
		return strings.TrimRight(strings.TrimSpace(p), "/"), nil
	}
	_ = timeout
	_ = rect
	return PickFileOrFolder("")
}

func CancelCaptureDrop() {}

func ClearCaptureDropFlags() {}
