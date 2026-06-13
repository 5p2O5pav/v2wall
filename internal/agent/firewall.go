package agent

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	ipsetName   = "v2wall_blacklist"
	iptablesChain = "INPUT"
)

// UpdateFirewall 根据最新的黑名单和白名单更新 ipset 和 iptables 规则
func UpdateFirewall(blacklist, whitelist []string) error {
	// 1. 确保 ipset 存在
	if err := ensureIpset(); err != nil {
		return err
	}

	// 2. 获取当前 ipset 中的 IP
	currentIPs, err := getIpsetMembers()
	if err != nil {
		return err
	}

	// 构建期望的黑名单集合，并排除白名单中的 IP
	desired := make(map[string]bool)
	for _, ip := range blacklist {
		if isIPInCIDRList(ip, whitelist) {
			continue
		}
		desired[ip] = true
	}

	// 3. 删除多余 IP
	for _, ip := range currentIPs {
		if !desired[ip] {
			log.Printf("[firewall] removing IP %s", ip)
			if err := runCmd("ipset", "del", ipsetName, ip); err != nil {
				log.Printf("[firewall] del %s failed: %v", ip, err)
			}
		}
	}

	// 4. 添加缺少的 IP
	for ip := range desired {
		if !contains(currentIPs, ip) {
			log.Printf("[firewall] adding IP %s", ip)
			if err := runCmd("ipset", "add", ipsetName, ip); err != nil {
				log.Printf("[firewall] add %s failed: %v", ip, err)
			}
		}
	}

	// 5. 确保 iptables 规则存在（如果不存在则插入）
	return ensureIptablesRule()
}

func ensureIpset() error {
	// 检查 ipset 是否已存在
	check := exec.Command("ipset", "list", ipsetName)
	if check.Run() == nil {
		return nil
	}
	// 创建 hash:ip 集合
	return runCmd("ipset", "create", ipsetName, "hash:ip", "timeout", "0")
}

func getIpsetMembers() ([]string, error) {
	cmd := exec.Command("ipset", "list", ipsetName)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var members []string
	inMembers := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "Members:" {
			inMembers = true
			continue
		}
		if inMembers && line != "" && !strings.HasPrefix(line, "Type") {
			members = append(members, line)
		}
	}
	return members, nil
}

func ensureIptablesRule() error {
	// 检查 INPUT 链中是否已有匹配 ipset 的规则
	check := exec.Command("iptables", "-C", iptablesChain, "-m", "set", "--match-set", ipsetName, "src", "-j", "DROP")
	if check.Run() == nil {
		return nil
	}
	// 添加规则到 INPUT 顶部
	return runCmd("iptables", "-I", iptablesChain, "1", "-m", "set", "--match-set", ipsetName, "src", "-j", "DROP")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %v, output: %s", name, args, err, string(output))
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isIPInCIDRList(ip string, cidrs []string) bool {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}
	for _, cidrStr := range cidrs {
		_, cidr, err := net.ParseCIDR(cidrStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ipAddr) {
			return true
		}
	}
	return false
}
