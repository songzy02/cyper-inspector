package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cyber-inspector/internal/bootstrap"
	"cyber-inspector/internal/config"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	AppName    = "Cyber Inspector"
	AppVersion = "2.0.0"
)

// @title Cyber Inspector API
// @version 2.0.0
// @description 智能分布式巡检监控系统
// @host localhost:8080
// @BasePath /
func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// 1. 先加载配置
	if err := config.Load(*configFile); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 2. 配置已就绪，再打印横幅
	printBanner()

	// 3. 启动应用
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("应用退出")
}

func printBanner() {
	banner := `
    ____      _       _ _     _____ _            _   
   / ___| ___| |_ ___| | |   |_   _| |__   ___  | |_ 
  | |    / _ \ __/ __| | |     | | | '_ \ / _ \ | __|
  | |___|  __/ || (__| | |     | | | | | |  __/ | |_ 
   \____|\___|\__\___|_|_|     |_| |_| |_|\___|  \__|
                                                      
    ` + fmt.Sprintf("v%s - 智能分布式巡检监控系统", AppVersion) + `
    ` + fmt.Sprintf("监听地址: http://localhost%s", config.Conf.Server.Listen) + `
    ` + fmt.Sprintf("环境模式: %s", config.Conf.App.Env) + `
    ` + fmt.Sprintf("巡检间隔: %v", config.Conf.Check.Interval) + `
    `
	fmt.Println(banner)
}
