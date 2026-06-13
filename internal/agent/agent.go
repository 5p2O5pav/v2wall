package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/config"
	"github.com/yourorg/v2wall/internal/honeypot"
	"github.com/yourorg/v2wall/internal/logwriter"
)

// Agent 被控端结构体
type Agent struct {
	cfg         *config.Config
	httpClient  *http.Client
	stopCh      chan struct{}
	logCollector *LogCollector
}

func New(cfg *config.Config) (*Agent, error) {
	return &Agent{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		stopCh:     make(chan struct{}),
	}, nil
}

func (a *Agent) Start() {
	log.Println("[agent] starting...")

	// 启动防火墙同步循环
	go a.syncLoop()

	// 如果开启诱饵，启动诱饵服务与日志收集
	if a.cfg.Agent.Honeypot.Enabled {
		a.logCollector = NewLogCollector()
		go a.startHoneypot()
		go a.reportLoop()
	}
}

func (a *Agent) Stop() {
	close(a.stopCh)
	// 清理防火墙？根据需求可保留规则
}

// syncLoop 定期同步黑名单并更新防火墙
func (a *Agent) syncLoop() {
	interval := time.Duration(a.cfg.Agent.SyncInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 首次启动立即同步一次
	a.syncFirewall()

	for {
		select {
		case <-ticker.C:
			a.syncFirewall()
			// 发送心跳
			a.sendHeartbeat()
		case <-a.stopCh:
			return
		}
	}
}

func (a *Agent) syncFirewall() {
	// 拉取黑名单
	blacklist, err := a.fetchBlacklist()
	if err != nil {
		log.Printf("[agent] fetch blacklist error: %v", err)
		return
	}
	// 拉取白名单
	whitelist, err := a.fetchWhitelist()
	if err != nil {
		log.Printf("[agent] fetch whitelist error: %v", err)
		return
	}

	// 合并本地白名单
	whitelist = append(whitelist, a.cfg.Agent.Whitelist...)

	// 更新防火墙
	if err := UpdateFirewall(blacklist, whitelist); err != nil {
		log.Printf("[agent] firewall update error: %v", err)
	}
}

func (a *Agent) fetchBlacklist() ([]string, error) {
	resp, err := a.httpClient.Get(a.masterURL("/api/v1/sync/blacklist"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Blacklist []string `json:"blacklist"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Blacklist, nil
}

func (a *Agent) fetchWhitelist() ([]string, error) {
	resp, err := a.httpClient.Get(a.masterURL("/api/v1/sync/whitelist"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Whitelist []string `json:"whitelist"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Whitelist, nil
}

func (a *Agent) sendHeartbeat() {
	payload := map[string]interface{}{
		"node_id":           getHostname(),
		"honeypot_enabled":  a.cfg.Agent.Honeypot.Enabled,
	}
	data, _ := json.Marshal(payload)
	_, _ = a.httpClient.Post(a.masterURL("/api/v1/sync/heartbeat"), "application/json", bytes.NewReader(data))
}

// startHoneypot 启动本地诱饵服务
func (a *Agent) startHoneypot() {
	// 诱饵不连接数据库，只收集日志到内存
	router := gin.Default()
	// 使用共享的诱饵处理器，但需要传入 nil db 和 nil searcher，并通过收集器捕获日志
	router.NoRoute(func(c *gin.Context) {
		// 捕获请求信息
		entry := captureEntry(c)
		a.logCollector.Add(entry)
		// 使用共享的响应逻辑（返回静态页面）
		honeypot.Handler(nil, nil, false, false)(c)
	})

	addr := fmt.Sprintf(":%d", a.cfg.Agent.Honeypot.Port)
	log.Printf("[agent] honeypot listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Printf("[agent] honeypot error: %v", err)
	}
}

// captureEntry 从 gin context 提取日志条目（复制自主控逻辑，不写库）
func captureEntry(c *gin.Context) logwriter.LogEntry {
	now := time.Now().UnixNano()
	ip := c.ClientIP()
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		ip = xri
	}
	return logwriter.LogEntry{
		IP:        ip,
		Time:      now,
		Method:    c.Request.Method,
		URL:       c.Request.URL.String(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	}
}

// reportLoop 定时上报日志
func (a *Agent) reportLoop() {
	interval := time.Duration(a.cfg.Agent.Honeypot.ReportInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			entries := a.logCollector.Flush()
			if len(entries) == 0 {
				continue
			}
			if err := a.reportLogs(entries); err != nil {
				log.Printf("[agent] report logs error: %v", err)
				// 失败时是否重新放回？简单起见丢弃，可根据需要保留
			}
		case <-a.stopCh:
			return
		}
	}
}

func (a *Agent) reportLogs(entries []logwriter.LogEntry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	resp, err := a.httpClient.Post(a.masterURL("/api/v1/sync/report"), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report failed: %s", string(body))
	}
	return nil
}

// masterURL 拼接带 token 的完整 URL
func (a *Agent) masterURL(path string) string {
	base := a.cfg.Agent.MasterURL
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%s%stoken=%s", base, path, sep, a.cfg.Agent.SyncToken)
}

func getHostname() string {
	name, _ := os.Hostname()
	return name
}
