package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"zsxyww.com/wts/config"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

var usage string = `
Usage: wtstool -a [action] -c [configure file path] \
`

func main() {
	cfg := config.Load()
	action := strings.Split(cfg.Actions, " ")
	var err error

	switch action[0] {
	case "set_wx_menu":
		// wtstool -a "set_wx_menu default menu.json" -c config.yaml
		err = setWXMenu(cfg, action[1], action[2])
	case "set_wx_tags":
		// wtstool -a "set_wx_tags [tagname]" -c config.yaml
		err = setWXTags(cfg, action[1])
	case "get_wx_tags":
		// wtstool -a "get_wx_tags" -c config.yaml
		err = getWXTags(cfg)
	case "get_wx_menu":

	case "gen_jwt_key":
		// wtstool -a "gen_jwt_key [OpenID] [sid] [access] [username] [avatar] [name] [expire]" -c config.yaml
		err = genJWTKey(cfg, action[1], action[2], action[3], action[4], action[5], action[6], action[7])
	case "change_wx_tag":
		// wtstool -a "change_wx_tag [OpenID] [tag]" -c config.yaml
		err = changeWXTag(cfg, action[1], action[2])
	default:
		fmt.Println("未知的指令，本程序用法见下：")
		fmt.Println(usage)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("执行", action[0], "时出现错误：", err)
		os.Exit(1)
	}
	fmt.Println("ok")
}

func genJWTKey(cfg *config.Config, openID, sid string, accessString string, username string, avatar string, name string, expireString string) error {
	expire, err := strconv.Atoi(expireString)
	if err != nil {
		return err
	}
	if !isValidAccess(accessString) {
		return fmt.Errorf("无效的访问权限: %s", accessString)
	}
	access := sqlc.WtsAccess(accessString)
	hutil.InitJWTKey(cfg.JWTKey)
	token, err := hutil.NewWtsJWT(openID, sid, access, username, avatar, name, expire)
	if err != nil {
		return err
	}

	fmt.Printf("生成的JWT令牌为：\n%s\n", token)
	return nil
}

func isValidAccess(access string) bool {
	return access == "dev" ||
		access == "chief" ||
		access == "api" ||
		access == "user" ||
		access == "unregistered" ||
		access == "formal-member" ||
		access == "informal-member" ||
		access == "pre-member" ||
		access == "group-leader"
}
