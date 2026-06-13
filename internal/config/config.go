package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Master  MasterConfig  `mapstructure:"master"`
	Agent   AgentConfig   `mapstructure:"agent"`
	Honeypot HoneypotConfig `mapstructure:"honeypot"`
	IPGeo   IPGeoConfig   `mapstructure:"ipgeo"`
}

type MasterConfig struct {
	AdminPath   string `mapstructure:"admin_path"`
	JWTSecret   string `mapstructure:"jwt_secret"`
	InitToken   string `mapstructure:"init_token"`
	SyncToken   string `mapstructure:"sync_token"`
	ListenPort  int    `mapstructure:"listen_port"`
	HoneypotPort int   `mapstructure:"honeypot_port"`
	DBPath      string `mapstructure:"db_path"`
	IPGeoV4     string `mapstructure:"ipgeo_v4_path"`
	IPGeoV6     string `mapstructure:"ipgeo_v6_path"`
	EnableIPv4  bool   `mapstructure:"enable_ipv4"`
	EnableIPv6  bool   `mapstructure:"enable_ipv6"`
}

type AgentConfig struct {
	MasterURL     string `mapstructure:"master_url"`
	SyncToken     string `mapstructure:"sync_token"`
	SyncInterval  int    `mapstructure:"sync_interval_sec"`
	Whitelist     []string `mapstructure:"local_whitelist"`
	Honeypot      HoneypotConfig `mapstructure:"honeypot"`
}

type HoneypotConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Port             int    `mapstructure:"port"`
	StaticDir        string `mapstructure:"static_dir"`        // 为空则用内嵌
	ReportInterval   int    `mapstructure:"report_interval_sec"`
}

type IPGeoConfig struct {
	EnableV4 bool   `mapstructure:"enable_v4"`
	DBPathV4 string `mapstructure:"db_path_v4"`
	EnableV6 bool   `mapstructure:"enable_v6"`
	DBPathV6 string `mapstructure:"db_path_v6"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
