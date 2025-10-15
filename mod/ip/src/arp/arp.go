package arp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Entry represents an ARP table entry.
type Entry struct {
	IP     string // IP address
	MAC    string // MAC address (hardware address)
	Device string // Interface/device name
}

// Table returns a slice of ARP entries from the system's ARP table.
// It supports Linux, Darwin (macOS), and Windows.
func Table() ([]Entry, error) {
	switch runtime.GOOS {
	case "linux":
		return tableLinux()
	case "darwin":
		return tableDarwin()
	case "windows":
		return tableWindows()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func tableLinux() ([]Entry, error) {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var entries []Entry
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		ip := fields[0]
		mac := fields[3]
		device := fields[5]
		entries = append(entries, Entry{IP: ip, MAC: mac, Device: device})
	}
	return entries, nil
}

func tableDarwin() ([]Entry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var entries []Entry
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		ip := strings.Trim(fields[1], "()")
		mac := fields[3]
		device := fields[5]
		if mac == "(incomplete)" {
			mac = ""
		}
		entries = append(entries, Entry{IP: ip, MAC: mac, Device: device})
	}
	return entries, nil
}

func tableWindows() ([]Entry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var entries []Entry
	var currentIfaceIP string
	var currentDevice string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Interface:") {
			parts := strings.Split(line, "---")
			if len(parts) == 2 {
				currentIfaceIP = strings.TrimSpace(parts[0][10:])
				currentDevice = getInterfaceNameByIP(currentIfaceIP)
			}
			continue
		}
		if strings.HasPrefix(line, "Internet Address") {
			continue // header
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			ip := fields[0]
			mac := fields[1]
			entries = append(entries, Entry{IP: ip, MAC: mac, Device: currentDevice})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// getInterfaceNameByIP finds the interface name associated with the given IP address.
func getInterfaceNameByIP(ipStr string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	targetIP := net.ParseIP(ipStr)
	if targetIP == nil {
		return ""
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.Equal(targetIP) {
				return iface.Name
			}
		}
	}
	return ""
}
