package config

import (
	"fmt"
	"strings"
)

// Config holds the configuration for the application.
type Config struct {
	Brand         string         `mapstructure:"Brand"`
	ListenPort    int            `mapstructure:"ListenPort"`
	DB            DBConfig       `mapstructure:"DB"`
	WX            WXConfig       `mapstructure:"WX"`
	FrontEnd      FrontEndConfig `mapstructure:"FrontEnd"`
	Debug         DebugConfig    `mapstructure:"Debug"`
	JWTKey        string         `mapstructure:"JWTKey"`
	FrontEndDir   string         `mapstructure:"FrontEndDir"`
	LogLevel      string         `mapstructure:"LogLevel"`
	JSONLogOutput bool           `mapstructure:"JSONLogOutput"`
	Actions       string         `mapstructure:"Actions"`
}

// DBConfig holds the database configuration.
type DBConfig struct {
	Type     string `mapstructure:"Type"`
	Path     string `mapstructure:"Path"`
	Port     string `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	Name     string `mapstructure:"Name"`
	SSL      bool   `mapstructure:"SSL"`
}

type WXConfig struct {
	AppID          string `mapstructure:"AppID"`
	AppSecret      string `mapstructure:"AppSecret"`
	Token          string `mapstructure:"Token"`
	EncodingAESKey string `mapstructure:"EncodingAESKey"`
	CallBackURL    string `mapstructure:"CallBackURL"`
}

type FrontEndConfig struct {
	OnAuthSuccess string `mapstructure:"OnAuthSuccess"`
}

type DebugConfig struct {
	APIVerbose     bool `mapstructure:"APIVerbose"`
	ProgramVerbose bool `mapstructure:"ProgramVerbose"`
	SkipJWTAuth    bool `mapstructure:"SkipJWTAuth"`
}

// String returns a string representation of the Config struct for debugging.
func (c *Config) String() string {
	if c.Actions != "" {
		return ""
	}
	var a strings.Builder
	a.WriteString("\n--- Loaded Configuration ---\n")
	a.WriteString(fmt.Sprintf("Brand: %s\n", c.Brand))
	a.WriteString(fmt.Sprintf("Listen Port: %d\n", c.ListenPort))
	a.WriteString("JWTKey: ***REDACTED***\n")
	a.WriteString(fmt.Sprintf("FrontEndDir: %s\n", c.FrontEndDir))
	a.WriteString(fmt.Sprintf("LogLevel: %s\n", c.LogLevel))
	a.WriteString(fmt.Sprintf("JSONLogOutput %t\n", c.JSONLogOutput))
	a.WriteString("Database:\n")
	a.WriteString(fmt.Sprintf("  Type: %s\n", c.DB.Type))
	a.WriteString(fmt.Sprintf("  Path: %s\n", c.DB.Path))
	a.WriteString(fmt.Sprintf("  Port: %s\n", c.DB.Port))
	a.WriteString(fmt.Sprintf("  User: %s\n", c.DB.User))
	a.WriteString("  Password: ***REDACTED***\n")
	a.WriteString(fmt.Sprintf("  Name: %s\n", c.DB.Name))
	a.WriteString(fmt.Sprintf("  SSL: %t\n", c.DB.SSL))
	a.WriteString("WeChat:\n")
	a.WriteString(fmt.Sprintf("  AppID: %s\n", c.WX.AppID))
	a.WriteString("  AppSecret: ***REDACTED***\n")
	a.WriteString(fmt.Sprintf("  Token: %s\n", c.WX.Token))
	a.WriteString("  EncodingAESKey: ***REDACTED***\n")
	a.WriteString("FrontEnd:\n")
	a.WriteString(fmt.Sprintf("  OnAuthSuccess: %s\n", c.FrontEnd.OnAuthSuccess))
	a.WriteString("Debug:\n")
	a.WriteString(fmt.Sprintf("  APIVerbose: %t\n", c.Debug.APIVerbose))
	a.WriteString(fmt.Sprintf("  ProgramVerbose: %t\n", c.Debug.ProgramVerbose))
	a.WriteString(fmt.Sprintf("  SkipJWTAuth: %t\n", c.Debug.SkipJWTAuth))
	a.WriteString("---------------------------\n")
	return a.String()
}
