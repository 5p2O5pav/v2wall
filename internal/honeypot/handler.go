package honeypot

import (
	"embed"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/db"
	"github.com/yourorg/v2wall/internal/ipgeo"
	"github.com/yourorg/v2wall/internal/logwriter"
)

//go:embed static/index.html static/404.html
var embeddedStatic embed.FS

// 获取客户端真实 IP
func getClientIP(c *gin.Context) string {
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return host
}

// Handler 返回一个 gin.HandlerFunc，用于处理诱饵请求。
// db, ipSearcher, enableV4, enableV6 由调用方提供。
func Handler(db *badger.DB, ipSearcher *ipgeo.Searcher, enableV4, enableV6 bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		now := time.Now().UnixNano()
		ua := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")

		entry := logwriter.LogEntry{
			IP:        clientIP,
			Time:      now,
			Method:    c.Request.Method,
			URL:       c.Request.URL.String(),
			UserAgent: ua,
			Referer:   referer,
		}

		// 仅在 db 不为 nil 时写入
		if db != nil {
			_ = logwriter.WriteLog(db, entry, ipSearcher, enableV4, enableV6)
		}

		// 响应静态页面
		if c.Request.URL.Path == "/" {
			c.FileFromFS("static/index.html", http.FS(embeddedStatic))
		} else {
			c.FileFromFS("static/404.html", http.FS(embeddedStatic))
		}
	}
}
