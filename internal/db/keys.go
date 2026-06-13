package db

import (
	"fmt"
	"strconv"
)

// 键前缀
const (
	PrefixLog      = "log:"
	PrefixIPInfo   = "ipinfo:"
	PrefixCounter  = "counter:"
	PrefixLastSeen = "last_seen:"
	PrefixBlacklistConfig = "blacklist:config"
	PrefixWhitelist       = "whitelist:"
	PrefixNode            = "node:"
	PrefixUser            = "user:"
)

func LogKey(ip string, tsNano int64) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", PrefixLog, ip, tsNano))
}

func IPInfoKey(ip string) []byte {
	return []byte(PrefixIPInfo + ip)
}

func CounterKey(ip string) []byte {
	return []byte(PrefixCounter + ip)
}

func LastSeenKey(ip string) []byte {
	return []byte(PrefixLastSeen + ip)
}

func BlacklistConfigKey() []byte {
	return []byte(PrefixBlacklistConfig)
}

func WhitelistKey(cidr string) []byte {
	return []byte(PrefixWhitelist + cidr)
}

func NodeKey(id string) []byte {
	return []byte(PrefixNode + id)
}

func UserKey(username string) []byte {
	return []byte(PrefixUser + username)
}

// 解析 log 键中的时间戳
func ParseLogKeyTs(key []byte) (int64, error) {
	// key 格式: log:<ip>:<ts_nano>
	s := string(key)
	lastColon := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			lastColon = i
			break
		}
	}
	if lastColon == -1 {
		return 0, fmt.Errorf("invalid log key: %s", s)
	}
	return strconv.ParseInt(s[lastColon+1:], 10, 64)
}
