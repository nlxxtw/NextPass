//go:build darwin

package folderdialog

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"b0go/apps/pass/lib/macdiag"
)

func runOSA(script string) (string, error) {
	// 只用 choose file/folder，不要 activate SystemUIServer / tell Finder（易触发「被拒绝」循环弹框）
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	macdiag.Infof("osascript script=%q out=%q err=%v", truncate(script, 120), truncate(s, 200), err)
	if err != nil {
		low := strings.ToLower(s + " " + err.Error())
		if strings.Contains(low, "user canceled") || strings.Contains(low, "user cancelled") {
			return "", nil
		}
		if strings.Contains(low, "not authorized") || strings.Contains(low, "不允许") ||
			strings.Contains(low, "被拒绝") || strings.Contains(low, "1002") ||
			strings.Contains(low, "(-1743)") || strings.Contains(low, "(-1728)") {
			return "", errors.New("系统拒绝弹出选择框。请到「系统设置 → 隐私与安全性 → 自动化」允许本程序/终端控制相关应用；或在访达选中文件后按 Option+Command+C 拷贝路径，粘贴到输入框")
		}
		if s != "" {
			return "", errors.New(s)
		}
		return "", err
	}
	return strings.TrimRight(s, "/"), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func escAS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// Pick 弹出 macOS 文件夹选择框（只弹一次）
func Pick(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件夹"
	}
	macdiag.Infof("Pick startDir=%q title=%q", startDir, title)
	// 不设 default location，避免无效路径导致失败后再弹第二次
	return runOSA(`POSIX path of (choose folder with prompt "` + escAS(title) + `")`)
}

// PickFile 弹出 macOS 文件选择框（只弹一次）
func PickFile(startDir, title string) (string, error) {
	if title == "" {
		title = "选择文件"
	}
	macdiag.Infof("PickFile startDir=%q title=%q", startDir, title)
	return runOSA(`POSIX path of (choose file with prompt "` + escAS(title) + `")`)
}

// PickFileOrFolder 用两个独立按钮流程，避免 display dialog + 二次权限
func PickFileOrFolder(startDir string) (string, error) {
	macdiag.Infof("PickFileOrFolder")
	// 直接选文件更常见；需要文件夹时用户点「选文件夹」按钮
	return PickFile(startDir, "选择要发给学生的文件（文件夹请点「选文件夹」）")
}

// FinderSelection 读取访达选中项；无权限时返回空字符串（不报错、不弹框）
func FinderSelection() (string, error) {
	script := `try
  tell application "Finder"
    if (count of (selection as list)) is 0 then return ""
    return POSIX path of (item 1 of (selection as alias list))
  end tell
on error
  return ""
end try`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	s := strings.TrimRight(strings.TrimSpace(string(out)), "/")
	macdiag.Infof("FinderSelection out=%q err=%v", s, err)
	if err != nil {
		return "", nil // 静默失败，避免前端连环弹 toast/对话框
	}
	return s, nil
}

func CaptureDrop(timeout time.Duration, rect DropRect) (string, error) {
	_ = timeout
	_ = rect
	// 拖放：只尝试静默读访达选中，绝不在这里再弹 choose 对话框（否则会「一直弹框」）
	p, _ := FinderSelection()
	if strings.TrimSpace(p) != "" {
		macdiag.Infof("CaptureDrop got=%q", p)
		return p, nil
	}
	macdiag.Infof("CaptureDrop empty — ask paste or click pick button")
	return "", errors.New("未能识别拖放路径。请点「选文件/选文件夹」，或访达选中后 Option+Command+C 粘贴路径")
}

func CancelCaptureDrop() {}

func ClearCaptureDropFlags() {}
