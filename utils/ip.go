package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
)

// GetIPVersion 判断IP版本
func GetIPVersion(ip string) string {
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		if parsedIP.To4() != nil {
			return "IPv4"
		}
		return "IPv6"
	}
	return "Unknown"
}

// GetLocalIP 获取本地IP
func GetLocalIP(targetIP string) (string, error) {
	// 根据目标IP类型选择合适的本地地址
	ip := net.ParseIP(targetIP)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", targetIP)
	}

	var localIP string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	isIPv4 := ip.To4() != nil
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if isIPv4 {
				if ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					localIP = ipnet.IP.String()
					break
				}
			} else {
				if ipnet.IP.To16() == nil && !ipnet.IP.IsLoopback() {
					localIP = ipnet.IP.String()
					break
				}
			}
		}
	}

	if localIP == "" {
		return "", fmt.Errorf("no suitable local IP found for target: %s", targetIP)
	}

	return localIP, nil
}

func GetHostName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

func GetNodeName() (string, error) {
	hostname, err := GetHostName()
	if err != nil {
		return "", err
	}

	pattern := `cdn([^-]*)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(hostname)
	if len(match) > 1 {
		nodeName := match[1]
		return nodeName, nil
	} else {
		return "", errors.New("未匹配到符合要求的内容")
	}

}
