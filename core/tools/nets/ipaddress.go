package nets

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

// isClashOrFakeIP 198.18.0.0/15 常被 Clash/Surge 等占用，不能当教室局域网地址
func isClashOrFakeIP(ip net.IP) bool {
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	return v4[0] == 198 && (v4[1] == 18 || v4[1] == 19)
}

func isUsableHostIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	if ip.To4() == nil {
		return false
	}
	if isClashOrFakeIP(ip) {
		return false
	}
	return true
}

// 教室场景优先：192.168 > 10.x > 172.16-31 > 其它
func privateClassScore(ip net.IP) int {
	v4 := ip.To4()
	if v4 == nil {
		return 99
	}
	if v4[0] == 192 && v4[1] == 168 {
		return 0
	}
	if v4[0] == 10 {
		return 1
	}
	if v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31 {
		return 2
	}
	if ip.IsPrivate() {
		return 3
	}
	return 8
}

func ifacePriority(name string) int {
	n := strings.ToLower(name)
	virtualHints := []string{
		"vethernet", "hyper-v", "vmware", "virtualbox", "vbox", "docker", "wsl",
		"bluetooth", "loopback", "isatap", "teredo", "meta", "tap", "tun", "utun",
		"ppp", "ipsec", "wireguard", "nordlynx", "clash", "awdl", "llw", "vpn",
	}
	for _, h := range virtualHints {
		if strings.Contains(n, h) {
			return 90
		}
	}
	switch {
	case strings.Contains(n, "wi-fi"), strings.Contains(n, "wifi"), strings.Contains(n, "wlan"),
		strings.Contains(n, "wireless"), strings.Contains(n, "无线"):
		return 0
	case strings.HasPrefix(n, "en"), // macOS
		strings.Contains(n, "ethernet"), strings.Contains(n, "eth"),
		strings.Contains(n, "本地"), strings.Contains(n, "以太网"):
		return 1
	default:
		return 40
	}
}

type lanCand struct {
	ip       string
	ifacePri int
	classPri int
}

// ListLANIPs 返回可用主机 IPv4（排除回环/链路本地/Clash 假 IP），教室网卡优先
func ListLANIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var list []lanCand
	seen := map[string]struct{}{}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		ipri := ifacePriority(iface.Name)
		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)
			if !ok || !isUsableHostIP(ipnet.IP) {
				continue
			}
			s := ipnet.IP.String()
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			list = append(list, lanCand{
				ip:       s,
				ifacePri: ipri,
				classPri: privateClassScore(ipnet.IP),
			})
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].ifacePri != list[j].ifacePri {
			return list[i].ifacePri < list[j].ifacePri
		}
		if list[i].classPri != list[j].classPri {
			return list[i].classPri < list[j].classPri
		}
		return list[i].ip < list[j].ip
	})
	out := make([]string, 0, len(list))
	for _, c := range list {
		out = append(out, c.ip)
	}
	return out
}

func udpOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		// 无外网时试常见网关探测（不要求通）
		conn, err = net.Dial("udp", "192.168.1.1:53")
		if err != nil {
			conn, err = net.Dial("udp", "10.0.0.1:53")
			if err != nil {
				return nil
			}
		}
	}
	defer conn.Close()
	la, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || la == nil {
		return nil
	}
	return la.IP
}

// GetOutBoundIP 获取教师机可供学生访问的局域网 IP（二维码/分享链接用）
// 优先真实私网；不会因 UDP 探测失败就直接给 127.0.0.1。
func GetOutBoundIP() string {
	// 1) UDP 私网出口（含经 VPN/Clash 选中的真实私网，排除 198.18 假段）
	if ip := udpOutboundIP(); isUsableHostIP(ip) && ip.IsPrivate() {
		return ip.String()
	}
	// 2) 网卡扫描：Wi‑Fi/以太网 + 192.168 优先
	if ips := ListLANIPs(); len(ips) > 0 {
		return ips[0]
	}
	// 3) 任意可用 UDP 出口（园区公网地址等）
	if ip := udpOutboundIP(); isUsableHostIP(ip) {
		return ip.String()
	}
	return "127.0.0.1"
}

// GetLocalIP 获取内网IP地址字符串
func GetLocalIP() string {
	return GetOutBoundIP()
}

// GetIPArr 用来获取有效的内网IP
func GetIPArr() ([]string, error) {
	ips := ListLANIPs()
	if len(ips) == 0 {
		if ip := GetOutBoundIP(); ip != "" && ip != "127.0.0.1" {
			return []string{ip}, nil
		}
		return nil, fmt.Errorf("未找到局域网 IP")
	}
	return ips, nil
}

// GetIPMap 用来获取有效的内网IP
func GetIPMap() (map[int]string, error) {
	ips, err := GetIPArr()
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(ips))
	for i, ip := range ips {
		m[i] = ip
	}
	return m, nil
}
