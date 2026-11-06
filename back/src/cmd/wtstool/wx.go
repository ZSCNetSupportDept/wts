package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/user"
	"github.com/silenceper/wechat/v2/util"
	"github.com/urfave/cli/v3"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/wechat"
)

func newWXCommand() *cli.Command {
	return &cli.Command{
		Name:   "wx",
		Usage:  "微信服务号相关功能",
		Action: groupAction,
		Commands: []*cli.Command{
			{
				Name:   "menu",
				Usage:  "管理微信菜单",
				Action: groupAction,
				Commands: []*cli.Command{
					{
						Name:  "set",
						Usage: "从JSON设置默认菜单或添加条件菜单",
						Arguments: []cli.Argument{
							&cli.StringArg{Name: "group", Required: true},
							&cli.StringArg{Name: "file", Required: true},
						},
						Action: setWXMenu,
					},
					{
						Name:  "get",
						Usage: "获取当前菜单（未实现）",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if err := rejectExtraArgs(cmd); err != nil {
								return err
							}
							return errors.New("微信菜单获取未实现")
						},
					},
				},
			},
			{
				Name:   "tag",
				Usage:  "管理微信标签",
				Action: groupAction,
				Commands: []*cli.Command{
					{
						Name:      "create",
						Usage:     "创建微信标签",
						Arguments: []cli.Argument{&cli.StringArg{Name: "name", Required: true}},
						Action:    createWXTag,
					},
					{
						Name:  "list",
						Usage: "列出微信标签",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: "text", Usage: "Output format: text or json"},
						},
						Action: listWXTags,
					},
				},
			},
			{
				Name:   "user",
				Usage:  "管理微信用户",
				Action: groupAction,
				Commands: []*cli.Command{
					{
						Name:  "tag",
						Usage: "设置用户的默认、操作员或管理员标签",
						Arguments: []cli.Argument{
							&cli.StringArg{Name: "openid", Required: true},
							&cli.StringArg{Name: "tag", Required: true},
						},
						Action: changeWXTag,
					},
				},
			},
		},
	}
}

func loadWXAccount(cmd *cli.Command) (*officialaccount.OfficialAccount, error) {
	cfg, err := loadToolConfig(cmd)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.WX.AppID) == "" || strings.TrimSpace(cfg.WX.AppSecret) == "" {
		return nil, errors.New("WX.AppID and WX.AppSecret must be configured")
	}
	return wechat.Setup(cfg), nil
}

func setWXMenu(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	group := cmd.StringArg("group")
	if group != "default" && group != "operator" && group != "admin" {
		return fmt.Errorf("unknown menu group %q", group)
	}
	path := cmd.StringArg("file")
	if path == "" {
		return errors.New("menu file path must not be empty")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read menu file: %w", err)
	}
	if !json.Valid(content) {
		return errors.New("menu file contains invalid JSON")
	}
	wx, err := loadWXAccount(cmd)
	if err != nil {
		return err
	}
	if group == "default" {
		err = wx.GetMenu().SetMenuByJSON(string(content))
	} else {
		err = wx.GetMenu().AddConditionalByJSON(string(content))
	}
	if err != nil {
		return wxOperationError("set menu", err)
	}
	_, err = fmt.Fprintln(cmd.Root().Writer, "Menu updated")
	return err
}

func createWXTag(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	name := cmd.StringArg("name")
	if strings.TrimSpace(name) == "" {
		return errors.New("tag name must not be empty")
	}
	wx, err := loadWXAccount(cmd)
	if err != nil {
		return err
	}
	tag, err := wx.GetUser().CreateTag(name)
	if err != nil {
		return wxOperationError("create tag", err)
	}
	if tag == nil {
		return errors.New("create tag: invalid WeChat response")
	}
	_, err = fmt.Fprintf(cmd.Root().Writer, "Created tag %d: %s\n", tag.ID, tag.Name)
	return err
}

func listWXTags(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	format := cmd.String("output")
	if format != "text" && format != "json" {
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
	wx, err := loadWXAccount(cmd)
	if err != nil {
		return err
	}
	tags, err := getWXTags(ctx, wx)
	if err != nil {
		return err
	}
	if format == "json" {
		return json.NewEncoder(cmd.Root().Writer).Encode(tags)
	}
	writer := tabwriter.NewWriter(cmd.Root().Writer, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tNAME\tCOUNT"); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := fmt.Fprintf(writer, "%d\t%s\t%d\n", tag.ID, tag.Name, tag.Count); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func getWXTags(ctx context.Context, wx *officialaccount.OfficialAccount) ([]*user.TagInfo, error) {
	token, err := wx.GetAccessTokenContext(ctx)
	if err != nil {
		return nil, wxOperationError("get access token", err)
	}
	query := url.Values{"access_token": []string{token}}
	response, err := util.HTTPGetContext(ctx, "https://api.weixin.qq.com/cgi-bin/tags/get?"+query.Encode())
	if err != nil {
		return nil, wxOperationError("list tags", err)
	}
	var result struct {
		ErrCode int64           `json:"errcode"`
		Tags    json.RawMessage `json:"tags"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("list tags: invalid WeChat response: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("list tags: WeChat error code %d", result.ErrCode)
	}
	if len(result.Tags) == 0 {
		return nil, errors.New("list tags: WeChat response is missing tags")
	}
	var tags []*user.TagInfo
	if err := json.Unmarshal(result.Tags, &tags); err != nil {
		return nil, fmt.Errorf("list tags: invalid WeChat response: %w", err)
	}
	if tags == nil {
		tags = []*user.TagInfo{}
	}
	for _, tag := range tags {
		if tag == nil {
			return nil, errors.New("list tags: invalid WeChat tag")
		}
	}
	return tags, nil
}

func changeWXTag(ctx context.Context, cmd *cli.Command) error {
	if err := rejectExtraArgs(cmd); err != nil {
		return err
	}
	openID := cmd.StringArg("openid")
	if strings.TrimSpace(openID) == "" {
		return errors.New("OpenID must not be empty")
	}
	tag := cmd.StringArg("tag")
	if tag != "default" && tag != "operator" && tag != "admin" {
		return fmt.Errorf("unknown tag %q", tag)
	}
	wx, err := loadWXAccount(cmd)
	if err != nil {
		return err
	}
	logicCtx := logic.Ctx{WX: wx}
	if err := logicCtx.ChangeUserTag(openID, tag); err != nil {
		return wxOperationError("change user tag", err)
	}
	_, err = fmt.Fprintln(cmd.Root().Writer, "User tag updated")
	return err
}

func wxOperationError(operation string, err error) error {
	var commonError *util.CommonError
	if errors.As(err, &commonError) {
		return fmt.Errorf("%s: WeChat error code %d", operation, commonError.ErrCode)
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return fmt.Errorf("%s: WeChat request failed", operation)
}
