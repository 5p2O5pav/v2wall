package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/v2wall/internal/agent"
	"github.com/yourorg/v2wall/internal/config"
)

func main() {
	cfgPath := "configs/agent.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 启动 agent 循环（阻断同步 + 可选诱饵）
	ag, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("create agent: %v", err)
	}
	ag.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ag.Stop()
}
