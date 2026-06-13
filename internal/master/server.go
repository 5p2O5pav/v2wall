package master

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/config"
	"github.com/yourorg/v2wall/internal/ipgeo"
)

// RegisterAdminRoutes 在传入的 gin.Engine 上注册所有管理端 API 和同步 API。
// 所有路由均挂载在 cfg.Master.AdminPath 路径组下。
// 需要 JWT 的接口使用 JWTAuthMiddleware，同步接口使用 SyncTokenMiddleware。
func RegisterAdminRoutes(r *gin.Engine, db *badger.DB, searcher *ipgeo.Searcher, cfg *config.Config) {
	adminPath := cfg.Master.AdminPath
	admin := r.Group(adminPath)

	// ---------- 初始化与登录（无认证） ----------
	admin.POST("/api/v1/init", handleInit(db, cfg))
	admin.POST("/api/v1/login", handleLogin(db, cfg))

	// ---------- 管理 API（需 JWT） ----------
	api := admin.Group("/api/v1")
	api.Use(JWTAuthMiddleware(cfg.Master.JWTSecret))
	{
		api.GET("/stats/ips", handleGetIPs(db))            // TODO
		api.GET("/logs", handleGetLogs(db))                // TODO
		api.GET("/ipinfo", handleGetIPInfo(db, searcher))  // TODO
		api.GET("/whitelist", handleGetWhitelist(db))      // TODO
		api.POST("/whitelist", handleAddWhitelist(db))     // TODO
		api.DELETE("/whitelist/:id", handleDeleteWhitelist(db)) // TODO
		api.GET("/config", handleGetConfig(db))            // TODO
		api.PUT("/config", handleUpdateConfig(db))         // TODO
		api.GET("/nodes", handleGetNodes(db))              // TODO
	}

	// ---------- 同步 API（需 Sync Token） ----------
	sync := admin.Group("/api/v1/sync")
	sync.Use(SyncTokenMiddleware(cfg.Master.SyncToken))
	{
		sync.GET("/blacklist", handleSyncBlacklist(db))      // TODO
		sync.GET("/whitelist", handleSyncWhitelist(db))      // TODO
		sync.POST("/heartbeat", handleSyncHeartbeat(db))     // TODO
		sync.POST("/report", handleReportLogs(db, searcher, cfg)) // TODO
	}
}
