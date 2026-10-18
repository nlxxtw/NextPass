package main

import (
	_ "b0go/apps/docs"
	"b0go/apps/panel"
	passapp "b0go/apps/pass"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"b0go/core/engine"
	"b0go/core/gateway"
	"b0go/core/tools/cmd"
	"b0go/core/tools/files"
	"b0go/core/tools/nets"

	"github.com/logrusorgru/aurora"
)

func chdirToExe() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return
	}
	_ = os.Chdir(filepath.Dir(exe))
}

func main() {
	runtime.LockOSThread()
	// Mac/Linux 从 Finder/访达打开时 cwd 常是 /，配置必须落在程序同目录
	chdirToExe()

	configFile := "config.ini"
	defaultConfig := "[gateway]\nListenAddr = \":8000\"\nDomain=\"\"\nLive = false\n\n[pass]\nPath = \"files\"\nCodeReadOnly = \"\"\nCodeReadWrite = \"admin\"\nLive = false\n"
	ok, _ := files.PathExists(configFile)
	if !ok {
		_ = os.WriteFile(configFile, []byte(defaultConfig), 0666)
		_ = os.MkdirAll("files", 0666)
	}

	// Windows：只用原生工具面板（无 WebView2、无黑框控制台）
	if runtime.GOOS == "windows" {
		if panel.RunDesktop() {
			return
		}
		// 面板失败不再静默退回黑框，避免「又变回以前那种」
		return
	}

	// 非 Windows：控制台自动启动
	gateway.AutoListen = true
	go func() {
		for i := 0; i < 80 && !gateway.IsRunning(); i++ {
			time.Sleep(100 * time.Millisecond)
		}
		port := gateway.PortOnly(gateway.CurrentAddr())
		if port == "" {
			port = "8000"
		}
		cwd, _ := os.Getwd()
		engine.Print(aurora.Black("--------------------------------------------"))
		engine.Print(aurora.BgGreen(aurora.Black("              NextPass 局域网作业助手              ")))
		engine.Print(aurora.BrightBlue("程序目录：" + cwd))
		engine.Print(aurora.BrightBlue("配置文件：" + filepath.Join(cwd, "config.ini")))
		if p := strings.TrimSpace(passapp.ConfigDataPath()); p != "" {
			engine.Print(aurora.BrightBlue("数据目录：" + p))
		}
		engine.Print(aurora.BrightBlue("端口：" + port))

		if strings.TrimSpace(engine.Domain) != "" {
			engine.Print(aurora.BrightBlue("学生访问： http://" + engine.Domain + "/"))
			engine.Print(aurora.BrightBlue("老师管理： http://" + engine.Domain + "/admin"))
		} else {
			shareIP := nets.GetOutBoundIP()
			ips := nets.ListLANIPs()
			if len(ips) == 0 && shareIP != "" && shareIP != "127.0.0.1" {
				ips = []string{shareIP}
			}
			printed := false
			for _, ip := range ips {
				if ip == "" || ip == "127.0.0.1" {
					continue
				}
				engine.Print(aurora.BrightBlue("学生访问： http://" + ip + ":" + port + "/"))
				printed = true
			}
			if !printed && shareIP != "" && shareIP != "127.0.0.1" {
				engine.Print(aurora.BrightBlue("学生访问： http://" + shareIP + ":" + port + "/"))
				printed = true
			}
			engine.Print(aurora.BrightBlue("本机管理： http://127.0.0.1:" + port + "/admin"))
			if !printed {
				engine.Print(aurora.Yellow("未识别到局域网 IP，请检查 Wi‑Fi/网线"))
			}
		}
		engine.Print(aurora.Green("改端口/密码：编辑同目录 config.ini 后重启"))
		engine.Print(aurora.Green("防火墙需允许本程序接受传入连接"))
		if p := strings.TrimSpace(passapp.ConfigDataPath()); p != "" {
			engine.Print(aurora.Yellow("Mac 排错日志：" + filepath.Join(p, "_nextpass", "mac_debug.log")))
		}
		engine.Print(aurora.Black("--------------------------------------------"))
		// 本机浏览器仍走 127.0.0.1（避免 Clash 假 IP 打不开）；二维码/分享链接用真实局域网 IP
		_ = cmd.Open("http://127.0.0.1:" + port + "/admin")
	}()
	time.Sleep(20 * time.Millisecond)
	_ = engine.Run(configFile)
	select {}
}
