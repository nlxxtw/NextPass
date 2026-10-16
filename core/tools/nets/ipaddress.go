package nets

import (
	"fmt"
	"net"
	"strings"
)

// GetOutBoundIP 使用 UDP 探测系统默认出口 IP（含 VPN/Clash 选中的地址）
func GetOutBoundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Printf("【错误】您的网络不通，请检查！ err: %s", err.Error())
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return strings.Split(localAddr.String(), ":")[0]
}

// GetLocalIP 获取内网IP地址字符串
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// GetIPArr 用来获取有效的内网IP
func GetIPArr() ([]string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("net.Interfaces failed, err: %s", err.Error())
	}
	var ip = make([]string, 0)
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil && ipnet.IP.String() != "" {
						ip = append(ip, ipnet.IP.String())
					}
				}
			}
		}
	}
	return ip, nil
}

// ListLANIPs 本机已启用网卡的 IPv4（不过滤 VPN/Clash）
func ListLANIPs() []string {
	ips, err := GetIPArr()
	if err != nil {
		return nil
	}
	return ips
}

// GetIPMap 用来获取有效的内网IP
func GetIPMap() (map[int]string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("net.Interfaces failed, err: %s", err.Error())
	}
	var ip = make(map[int]string)
	var j = 0
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil && ipnet.IP.String() != "" {
						ip[j] = ipnet.IP.String()
						j++
					}
				}
			}
		}
	}
	return ip, nil
}
