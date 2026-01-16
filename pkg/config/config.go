package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Config Portal 配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Billing  BillingConfig  `mapstructure:"billing"`
	MainSite MainSiteConfig `mapstructure:"main_site"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release
	Host string `mapstructure:"host"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	CookieName string `mapstructure:"cookie_name"` // 根据环境自动设置
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level  string           `mapstructure:"level"`  // debug, info, warn, error
	Output string           `mapstructure:"output"` // console, file, both
	File   LoggerFileConfig `mapstructure:"file"`
}

// LoggerFileConfig 日志文件配置
type LoggerFileConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"` // 天
	Compress   bool   `mapstructure:"compress"`
}

// BillingConfig 计费配置
type BillingConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	IntervalSeconds int  `mapstructure:"interval_seconds"` // 计费间隔(秒)
}

// MainSiteConfig 主站配置
type MainSiteConfig struct {
	URL    string `mapstructure:"url"`     // 主站地址
	APIURL string `mapstructure:"api_url"` // API 地址
}

// Load 加载配置
func Load() error {
	// 设置配置文件路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 允许环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 根据环境设置 Cookie 名称
	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	if cfg.JWT.CookieName == "" {
		if env == "production" {
			cfg.JWT.CookieName = "token"
		} else {
			cfg.JWT.CookieName = "test_token"
		}
	}

	// 设置全局配置
	GlobalConfig = &cfg

	return nil
}

// GetCookieName 获取 Cookie 名称
func GetCookieName() string {
	if GlobalConfig != nil && GlobalConfig.JWT.CookieName != "" {
		return GlobalConfig.JWT.CookieName
	}

	env := os.Getenv("ENV")
	if env == "production" {
		return "token"
	}
	return "test_token"
}

// IsAdminEmail 检查是否为管理员邮箱
func IsAdminEmail(email string) bool {
	// TODO: 从配置或数据库读取管理员邮箱列表
	adminEmails := []string{
		"admin@wavespeed.ai",
		"shanliu@wavespeed.ai",
	}

	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			return true
		}
	}
	return false
}
