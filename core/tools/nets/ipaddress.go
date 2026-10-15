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

func isGoodLAN(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	if isClashOrFakeIP(ip) {
		return false
	}
	return ip.IsPrivate()
}

func ifacePriority(name string) int {
	n := strings.ToLower(name)
	switch {
	case strings.HasPrefix(n, "en"): // macOS Wi‑Fi / 有线
		return 0
	case strings.HasPrefix(n, "eth"), strings.HasPrefix(n, "wlan"):
		return 1
	case strings.HasPrefix(n, "wl"):
		return 2
	case strings.Contains(n, "utun"), strings.Contains(n, "tun"), strings.Contains(n, "ppp"),
		strings.Contains(n, "ipsec"), strings.HasPrefix(n, "awdl"), strings.HasPrefix(n, "llw"):
		return 90
	default:
		return 50
	}
}

type lanCand struct {
	ip   string
	prio int
}

// ListLANIPs 返回可用的局域网 IPv4（已排除回环、链路本地、Clash 假 IP）
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
		prio := ifacePriority(iface.Name)
		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)
			if !ok || !isGoodLAN(ipnet.IP) {
				continue
			}
			s := ipnet.IP.String()
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			list = append(list, lanCand{ip: s, prio: prio})
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].prio != list[j].prio {
			return list[i].prio < list[j].prio
		}
		return list[i].ip < list[j].ip
	})
	out := make([]string, 0, len(list))
	for _, c := range list {
		out = append(out, c.ip)
	}
	return out
}

// GetOutBoundIP 优先返回真实局域网 IP（避开 VPN/Clash 假地址）
func GetOutBoundIP() string {
	if ips := ListLANIPs(); len(ips) > 0 {
		return ips[0]
	}
	// 回退：UDP 探测，但仍拒绝假 IP
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP
	if isGoodLAN(ip) {
		return ip.String()
	}
	if s := ip.String(); s != "" && !ip.IsLoopback() && !isClashOrFakeIP(ip) {
		return s
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
		return nil, fmt.Errorf("未找到局域网 IP")
	}
	return ips, nil
}

// GetIPMap 用来获取有效的内网IP
func GetIPMap() (map[int]string, error) {
	ips := ListLANIPs()
	if len(ips) == 0 {
		return nil, fmt.Errorf("未找到局域网 IP")
	}
	m := make(map[int]string, len(ips))
	for i, ip := range ips {
		m[i] = ip
	}
	return m, nil
}
