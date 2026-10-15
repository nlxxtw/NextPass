//go:build !windows
// +build !windows

package cmd

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var commands = map[string]string{
	"darwin": "open",
	"linux":  "xdg-open",
}

// Open 打开浏览器或本地路径
func Open(uri string) error {
	return Reveal(uri)
}

// Reveal 打开本地路径（非 Windows）
func Reveal(uri string) error {
	uri = strings.TrimSpace(uri)
	run, ok := commands[runtime.GOOS]
	if !ok {
		log.Printf("don't know how to open things on %s platform", runtime.GOOS)
		return nil
	}
	uri = strings.ReplaceAll(uri, "\\", "/")
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// 文件：在访达中显示并选中；文件夹：直接打开
		if st, err := os.Stat(uri); err == nil && !st.IsDir() {
			cmd = exec.Command(run, "-R", uri)
		} else {
			cmd = exec.Command(run, uri)
		}
	} else {
		cmd = exec.Command(run, uri)
	}
	log.Println("[CommandAs]", cmd.Args)
	return cmd.Start()
}
