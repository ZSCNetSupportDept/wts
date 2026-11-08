package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func registerFlags(flags *pflag.FlagSet) {
	flags.StringP("config", "c", "", "Path to config file")
	flags.String("Brand", "Zhongshan College Network HelpDesk Backend", "Brand")
	flags.Int("ListenPort", 25005, "Port to listen on")
	flags.String("DB.Type", "PostgreSQL", "Database type")
	flags.String("DB.Path", "127.0.0.1", "Database path")
	flags.String("DB.Port", "5432", "Database port")
	flags.String("DB.User", "", "Database user")
	flags.String("DB.Password", "", "Database password")
	flags.String("DB.Name", "", "Database name")
	flags.Bool("DB.SSL", false, "Enable SSL for database connection")
	flags.String("WX.AppID", "", "WeChat AppID")
	flags.String("WX.AppSecret", "", "WeChat AppSecret")
	flags.String("WX.Token", "", "WeChat Token")
	flags.String("WX.EncodingAESKey", "", "WeChat EncodingAESKey")
	flags.String("WX.CallBackURL", "", "WeChat CallBackURL")
	flags.Bool("WX.AIReply", false, "Transfer WeChat messages to AI replies")
	flags.String("QQBot.AppID", "", "QQBot AppID")
	flags.String("QQBot.AppSecret", "", "QQBot AppSecret")
	flags.Bool("QQBot.Enabled", false, "Enable QQBot integration")
	flags.String("JWTKey", "", "JWT signing key")
	flags.String("FrontEndDir", "", "Where to found FrontEnd Files")
	flags.String("FrontEnd.OnAuthSuccess", "/auth_success.html", "FrontEnd URL to redirect to on auth success")
	flags.String("FrontEnd.PublicURL", "https://wwbx.davisye.cn", "Public HTTPS frontend URL used in WeChat notification links")
	flags.String("LogLevel", "info", "Log level: debug, info, warn, error, panic, fatal")
	flags.Bool("Debug.APIVerbose", false, "Enable verbose API logging")
	flags.Bool("Debug.ProgramVerbose", false, "Enable verbose program logging")
	flags.Bool("Debug.SkipJWTAuth", false, "Skip JWT authentication (for debugging only)")
	flags.Bool("JSONLogOutput", false, "Output logs in JSON format")
	flags.StringP("Actions", "a", "", "wtstool only,actions")
}

func Load() *Config {
	registerFlags(pflag.CommandLine)
	pflag.Parse()
	configPath, _ := pflag.CommandLine.GetString("config")

	if configPath == "" {
		fmt.Println("Error: config file path must be provided via --config or -c")
		os.Exit(1)
	}

	c, err := load(configPath, pflag.CommandLine)
	if err != nil {
		panic(err)
	}
	if c.Debug.ProgramVerbose {
		fmt.Println(c)
	}
	return c
}

func LoadFile(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, errors.New("config file path must be provided")
	}
	flags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	registerFlags(flags)
	return load(configPath, flags)
}

func load(configPath string, flags *pflag.FlagSet) (*Config, error) {
	v := viper.New()
	if err := v.BindPFlags(flags); err != nil {
		return nil, err
	}
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file %q: %w", configPath, err)
	}

	v.SetEnvPrefix("WTS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return &c, nil
}
