package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/cleanup"
	"github.com/yourorg/v2wall/internal/config"
	"github.com/yourorg/v2wall/internal/db"
	"github.com/yourorg/v2wall/internal/honeypot"
	"github.com/yourorg/v2wall/internal/ipgeo"
	"github.com/yourorg/v2wall/internal/master"
)

//go:embed web/dist
var adminDist embed.FS

func main() {
	cfgPath := "configs/master.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	bdb, err := db.OpenDB(cfg.Master.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer bdb.Close()

	// 注册 counter merge 操作
	db.RegisterCounterMerge(bdb)

	// 初始化 IP 地理查询
	ipSearcher, err := ipgeo.NewSearcher(
		cfg.Master.IPGeoV4,
		cfg.Master.IPGeoV6,
		cfg.Master.EnableIPv4,
		cfg.Master.EnableIPv6,
	)
	if err != nil {
		log.Fatalf("init ipgeo: %v", err)
	}
	defer ipSearcher.Close()

	// 启动过期清理
	janitor := cleanup.NewJanitor(bdb, 30)
	go janitor.Run(6 * time.Hour)
	defer janitor.Stop()

	// ---------- 管理 API 服务 ----------
	adminRouter := gin.Default()

	// 挂载前端静态资源（必须在注册 API 路由前挂载，否则可能被 API 路由拦截）
	adminRouter.Use(static.Serve("/", static.EmbedFolder(adminDist, "web/dist")))

	// 注册管理 API 路由
	master.RegisterAdminRoutes(adminRouter, bdb, ipSearcher, cfg)

	adminAddr := fmt.Sprintf(":%d", cfg.Master.ListenPort)
	go func() {
		log.Printf("Admin server starting on %s", adminAddr)
		if err := adminRouter.Run(adminAddr); err != nil {
			log.Fatalf("admin server: %v", err)
		}
	}()

	// ---------- 诱饵服务 ----------
	honeypotRouter := gin.Default()
	honeypotRouter.NoRoute(honeypot.Handler(bdb, ipSearcher, cfg.Master.EnableIPv4, cfg.Master.EnableIPv6))
	honeypotAddr := fmt.Sprintf(":%d", cfg.Master.HoneypotPort)
	go func() {
		log.Printf("Honeypot server starting on %s", honeypotAddr)
		if err := honeypotRouter.Run(honeypotAddr); err != nil {
			log.Fatalf("honeypot server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
