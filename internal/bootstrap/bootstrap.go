package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cyber-inspector/internal/agent"
	"cyber-inspector/internal/config"
	"cyber-inspector/internal/handler"
	"cyber-inspector/internal/model"
	"cyber-inspector/internal/repository"
	"cyber-inspector/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run 启动应用
func Run() error {
	// 初始化数据库
	db, err := initDatabase()
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化仓库
	repo := repository.New(db)

	// 初始化管理员账户
	if err := repo.InitAdminUser("admin", "admin123"); err != nil {
		log.Printf("初始化管理员账户失败: %v", err)
	}

	// 创建 HTTP 客户端
	httpClient := agent.NewClient()

	// 创建巡检服务
	checker := service.NewChecker(repo, httpClient)

	// 启动巡检服务
	checker.Start()
	defer checker.Stop()

	// 创建 Gin 引擎
	gin.SetMode(getGinMode())
	engine := gin.New()

	// 配置中间件
	setupMiddleware(engine)

	// 注册路由
	setupRoutes(engine, repo, checker, db)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         config.Conf.Server.Listen,
		Handler:      engine,
		ReadTimeout:  config.Conf.Server.ReadTimeout,
		WriteTimeout: config.Conf.Server.WriteTimeout,
		IdleTimeout:  config.Conf.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		log.Printf("启动 HTTP 服务器: %s", config.Conf.Server.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("接收到退出信号，开始优雅关闭...")

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP 服务器关闭失败: %v", err)
	}

	log.Println("应用已优雅关闭")
	return nil
}

// initDatabase 初始化数据库
func initDatabase() (*gorm.DB, error) {
	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getGormLogMode()),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(config.Conf.MySQL.DSN), gormConfig)
	if err != nil {
		return nil, err
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(config.Conf.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Conf.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.Conf.MySQL.ConnMaxLifetime)

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("数据库连接成功")
	return db, nil
}

// autoMigrate 自动迁移数据库
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Agent{},
		&model.Inspection{},
		&model.Alert{},
		&model.LoginLog{},
	)
}

// setupMiddleware 配置中间件
func setupMiddleware(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(corsMiddleware())
}

// setupRoutes 注册路由
func setupRoutes(engine *gin.Engine, repo *repository.Repository, checker *service.Checker, db *gorm.DB) {
	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": config.Conf.App.Version,
		})
	})
	//就绪探针
	engine.GET("/readyz", func(c *gin.Context) {
		if err := readyCheck(db); err != nil {
			c.JSON(503, gin.H{"status": "not ready", "reason": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	// API 路由组
	api := engine.Group("/api")
	{
		// 认证相关
		api.POST("/auth/login", handler.Login(repo))
		api.POST("/auth/logout", handler.Logout())

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(handler.AuthMiddleware())
		{
			// Agent 管理
			auth.GET("/agents", handler.ListAgents(repo))
			auth.POST("/agents", handler.CreateAgent(repo))
			auth.PUT("/agents/:id", handler.UpdateAgent(repo))
			auth.DELETE("/agents/:id", handler.DeleteAgent(repo))
			auth.PUT("/agents/:id/interval", handler.UpdateInterval(repo))

			// 巡检相关
			auth.POST("/trigger", handler.TriggerCheck(checker))
			auth.GET("/status", handler.GetStatus(checker))
		}
	}

	// 前端页面
	engine.LoadHTMLGlob("web/templates/*")
	engine.Static("/static", "./web/static")
	engine.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	// 404 处理
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "接口不存在"})
	})

}

// readyCheck 返回 nil 表示已就绪
// readyCheck 返回 nil 表示已就绪
func readyCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	return nil
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getGinMode 获取Gin模式
func getGinMode() string {
	if config.Conf.IsProduction() {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}

// getGormLogMode 获取GORM日志模式
func getGormLogMode() logger.LogLevel {
	switch config.Conf.Log.Level {
	case "debug":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}
