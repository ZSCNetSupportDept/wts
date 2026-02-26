package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Load() *Config {
	v := viper.New()

	// Bind command line flags
	pflag.StringP("config", "c", "", "Path to config file")
	pflag.String("Brand", "Zhongshan College Network HelpDesk Backend", "Brand")
	pflag.Int("ListenPort", 25005, "Port to listen on")
	pflag.String("DB.Type", "PostgreSQL", "Database type")
	pflag.String("DB.Path", "127.0.0.1", "Database path")
	pflag.String("DB.Port", "5432", "Database port")
	pflag.String("DB.User", "", "Database user")
	pflag.String("DB.Password", "", "Database password")
	pflag.String("DB.Name", "", "Database name")
	pflag.Bool("DB.SSL", false, "Enable SSL for database connection")
	pflag.String("WX.AppID", "", "WeChat AppID")
	pflag.String("WX.AppSecret", "", "WeChat AppSecret")
	pflag.String("WX.Token", "", "WeChat Token")
	pflag.String("WX.EncodingAESKey", "", "WeChat EncodingAESKey")
	pflag.String("WX.CallBackURL", "", "WeChat CallBackURL")
	pflag.String("JWTKey", "", "JWT signing key")
	pflag.String("FrontEndDir", "", "Where to found FrontEnd Files")
	pflag.String("FrontEnd.OnAuthSuccess", "/auth_success.html", "FrontEnd URL to redirect to on auth success")
	pflag.String("LogLevel", "info", "Log level: debug, info, warn, error, panic, fatal")
	pflag.Bool("Debug.APIVerbose", false, "Enable verbose API logging")
	pflag.Bool("Debug.ProgramVerbose", false, "Enable verbose program logging")
	pflag.Bool("Debug.SkipJWTAuth", false, "Skip JWT authentication (for debugging only)")
	pflag.Bool("JSONLogOutput", false, "Output logs in JSON format")
	pflag.StringP("Actions", "a", "", "wtstool only,actions")
	pflag.Parse()

	// Check for config file path
	configPath, _ := pflag.CommandLine.GetString("config")

	if configPath == "" {
		fmt.Println("Error: config file path must be provided via --config or -c")
		os.Exit(1)
	}

	v.BindPFlags(pflag.CommandLine)

	// Load from config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		// Handle errors reading the config file
		// We can ignore "not found" error, as we have other config sources
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	// Load from environment variables
	v.SetEnvPrefix("WTS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		panic(err)
	}

	if c.Debug.ProgramVerbose {
		fmt.Println(&c)
	}

	return &c
}
