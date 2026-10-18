package macdiag

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	mu   sync.Mutex
	path string
)

// Setup 设置 Mac 调试日志路径（如 …/_nextpass/mac_debug.log）
func Setup(logPath string) {
	if runtime.GOOS != "darwin" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	path = logPath
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	Infof("==== mac_debug start pid=%d ====", os.Getpid())
}

// Path 当前日志文件
func Path() string {
	mu.Lock()
	defer mu.Unlock()
	return path
}

// Infof 写一行到终端 + mac_debug.log（仅 darwin 写文件；其它系统只打标准日志）
func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := time.Now().Format("2006-01-02 15:04:05") + " " + msg
	log.Println("[macdiag]", msg)
	if runtime.GOOS != "darwin" {
		return
	}
	mu.Lock()
	p := path
	mu.Unlock()
	if p == "" {
		return
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line + "\n")
	_ = f.Close()
}
