package main

import (
	"context"
	"fmt"

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
