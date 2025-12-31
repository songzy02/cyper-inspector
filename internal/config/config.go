// file: internal/config/config.go
package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// 全局配置变量
var Conf *Config

// Config 总配置
type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Check  CheckConfig  `mapstructure:"check"`
	Alert  AlertConfig  `mapstructure:"alert"`
	Mail   MailConfig   `mapstructure:"mail"`
	LLM    LLMConfig    `mapstructure:"llm"`
	Log    LogConfig    `mapstructure:"log"`
	JWT    JWTConfig    `mapstructure:"jwt"` // <-- 新增
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Secret  string `mapstructure:"secret"`
	Env     string `mapstructure:"env"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Listen       string        `mapstructure:"listen"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// CheckConfig 巡检配置
type CheckConfig struct {
	Interval      time.Duration `mapstructure:"interval"`
	Timeout       time.Duration `mapstructure:"timeout"`
	MaxConcurrent int           `mapstructure:"max_concurrent"`
	RetryTimes    int           `mapstructure:"retry_times"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Cooldown  time.Duration `mapstructure:"cooldown"`
	Threshold struct {
		CPU     float64 `mapstructure:"cpu"`
		Memory  float64 `mapstructure:"memory"`
		Disk    float64 `mapstructure:"disk"`
		LoadAvg float64 `mapstructure:"load_avg"`
	} `mapstructure:"threshold"`
}

// MailConfig 邮件配置
type MailConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	User          string `mapstructure:"user"`
	Pass          string `mapstructure:"pass"`
	To            string `mapstructure:"to"`
	SubjectPrefix string `mapstructure:"subject_prefix"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	APIURL      string        `mapstructure:"api_url"`
	Model       string        `mapstructure:"model"`
	Temperature float64       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
	MaxTokens   int           `mapstructure:"max_tokens"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    string `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     string `mapstructure:"max_age"`
}

// JWTConfig JWT 配置  <-- 新增
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// Load 加载配置
func Load(configFile string) error {
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("配置文件未找到，使用默认配置: %s", configFile)
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	Conf = &Config{}
	if err := v.Unmarshal(Conf); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	return validateConfig()
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "Cyber Inspector")
	v.SetDefault("app.version", "2.0.0")
	v.SetDefault("app.secret", "cyber-inspector-secret")
	v.SetDefault("app.env", "development")

	v.SetDefault("server.listen", ":8080")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")

	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.conn_max_lifetime", "5m")

	v.SetDefault("check.interval", "5m")
	v.SetDefault("check.timeout", "30s")
	v.SetDefault("check.max_concurrent", 10)
	v.SetDefault("check.retry_times", 3)

	v.SetDefault("alert.enabled", true)
	v.SetDefault("alert.cooldown", "5m")
	v.SetDefault("alert.threshold.cpu", 85.0)
	v.SetDefault("alert.threshold.memory", 90.0)
	v.SetDefault("alert.threshold.disk", 90.0)
	v.SetDefault("alert.threshold.load_avg", 5.0)

	v.SetDefault("mail.enabled", false)
	v.SetDefault("mail.port", 994)
	v.SetDefault("mail.subject_prefix", "[Cyber Inspector]")

	v.SetDefault("llm.enabled", false)
	v.SetDefault("llm.temperature", 0.0)
	v.SetDefault("llm.timeout", "30s")
	v.SetDefault("llm.max_tokens", 2000)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", "100MB")
	v.SetDefault("log.max_backups", 5)
	v.SetDefault("log.max_age", "30d")

	// JWT 默认值
	v.SetDefault("jwt.secret", "change-me")
	v.SetDefault("jwt.expire_hours", 24)
}

// validateConfig 基础校验
func validateConfig() error {
	if Conf.MySQL.DSN == "" {
		return fmt.Errorf("MySQL DSN 不能为空")
	}
	if Conf.Server.Listen == "" {
		return fmt.Errorf("服务器监听地址不能为空")
	}
	return nil
}

// IsDevelopment 是否开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
