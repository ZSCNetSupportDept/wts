package main

import (
	"errors"
	"os"

	"zsxyww.com/wts/config"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/wechat"
)

func setDefaultWXMenu(cfg *config.Config, file string) error {

	wx := wechat.Setup(cfg)
	menu := wx.GetMenu()

	if file == "" {
		err := errors.New("no menu file selected")
		return err
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = menu.SetMenuByJSON(string(content))
	if err != nil {
		return err
	}
	return nil
}

func setWXTags(cfg *config.Config, newTag string) error {
	wx := wechat.Setup(cfg)
	u := wx.GetUser()

	if newTag == "" {
		err := errors.New("no tag selected")
		return err
	}
	tag, err := u.CreateTag(newTag)
	if err != nil {
		return err
	}
	println(tag, " Created tag with ID:", tag.ID)
	return nil
}

func getWXTags(cfg *config.Config) error {
	wx := wechat.Setup(cfg)
	u := wx.GetUser()

	tags, err := u.GetTag()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		println("Tag ID:", tag.ID, "Name:", tag.Name, "Count:", tag.Count)
	}
	return nil
}

func setConditionalMenu(cfg *config.Config, file string) error {

	wx := wechat.Setup(cfg)
	menu := wx.GetMenu()

	if file == "" {
		err := errors.New("no menu file selected")
		return err
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = menu.AddConditionalByJSON(string(content))
	if err != nil {
		return err
	}
	return nil
}

func setWXMenu(cfg *config.Config, group string, file string) error {
	switch group {
	case "default":
		return setDefaultWXMenu(cfg, file)
	case "operator":
		return setConditionalMenu(cfg, file)
	case "admin":
		return setConditionalMenu(cfg, file)
	default:
		return errors.New("unknown menu group: " + group)
	}
}

func changeWXTag(cfg *config.Config, openID string, tag string) error {
	if tag != "default" && tag != "operator" && tag != "admin" {
		return errors.New("unknown tag: " + tag)
	}

	ctx := logic.Ctx{
		WX: wechat.Setup(cfg),
	}
	return ctx.ChangeUserTag(openID, tag)

}
