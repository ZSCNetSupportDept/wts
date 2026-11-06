package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/urfave/cli/v3"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

func newJWTCommand() *cli.Command {
	return &cli.Command{
		Name:   "jwt",
		Usage:  "JWT相关功能",
		Action: groupAction,
		Commands: []*cli.Command{
			{
				Name:  "issue",
				Usage: "签发JWT",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "openid", Usage: "WeChat OpenID", Required: true},
					&cli.StringFlag{Name: "access", Usage: "Access role", Required: true},
					&cli.IntFlag{Name: "expire-minutes", Usage: "Validity in minutes", Required: true},
					&cli.StringFlag{Name: "sid", Usage: "Student ID"},
					&cli.StringFlag{Name: "username", Usage: "Username"},
					&cli.StringFlag{Name: "avatar", Usage: "Avatar URL"},
					&cli.StringFlag{Name: "name", Usage: "Name"},
				},
				Action: issueJWT,
			},
			{
				Name:      "parse",
				Usage:     "解析并验证JWT",
				Arguments: []cli.Argument{&cli.StringArg{Name: "token", Required: true}},
				Action:    parseJWT,
			},
		},
	}
}

func issueJWT(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	access := sqlc.WtsAccess(cmd.String("access"))
	if !isValidAccess(access) {
		return fmt.Errorf("invalid access role %q", access)
	}
	cfg, err := loadToolConfig(cmd)
	if err != nil {
		return err
	}
	if cfg.JWTKey == "" {
		return fmt.Errorf("JWTKey must not be empty")
	}
	hutil.InitJWTKey(cfg.JWTKey)
	token, err := hutil.NewWtsJWT(
		cmd.String("openid"),
		cmd.String("sid"),
		access,
		cmd.String("username"),
		cmd.String("avatar"),
		cmd.String("name"),
		cmd.Int("expire-minutes"),
	)
	if err != nil {
		return fmt.Errorf("issue JWT: %w", err)
	}
	_, err = fmt.Fprintln(cmd.Root().Writer, token)
	return err
}

func parseJWT(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	tokenString := cmd.StringArg("token")
	if tokenString == "" {
		return fmt.Errorf("token must not be empty")
	}
	cfg, err := loadToolConfig(cmd)
	if err != nil {
		return err
	}
	if cfg.JWTKey == "" {
		return fmt.Errorf("JWTKey must not be empty")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT: expected three segments")
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("decode JWT header: %w", err)
	}
	if !json.Valid(header) {
		return fmt.Errorf("decode JWT header: invalid JSON")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decode JWT payload: %w", err)
	}
	if !json.Valid(payload) {
		return fmt.Errorf("decode JWT payload: invalid JSON")
	}

	_, verifyErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithoutClaimsValidation(), jwt.WithJSONNumber())
	result := struct {
		Header   json.RawMessage `json:"header"`
		Payload  json.RawMessage `json:"payload"`
		Verified bool            `json:"verified"`
	}{Header: header, Payload: payload, Verified: verifyErr == nil}
	encoder := json.NewEncoder(cmd.Root().Writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("write JWT: %w", err)
	}
	if verifyErr != nil {
		return fmt.Errorf("verify JWT: %w", verifyErr)
	}
	return nil
}

func isValidAccess(access sqlc.WtsAccess) bool {
	switch access {
	case sqlc.WtsAccessDev, sqlc.WtsAccessChief, sqlc.WtsAccessApi,
		sqlc.WtsAccessUser, sqlc.WtsAccessUnregistered, sqlc.WtsAccessFormalMember,
		sqlc.WtsAccessInformalMember, sqlc.WtsAccessPreMember, sqlc.WtsAccessGroupLeader:
		return true
	default:
		return false
	}
}
