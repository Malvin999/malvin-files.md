package userconfig

import (
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/constants"
	"zakirullin/stuffbot/pkg/tg"
)

type QuickPanelBtn struct {
	Cmd         string
	CmdType     string
	Emoji       string
	Description string
}

var QuickPanelAvailableBtns = []QuickPanelBtn{
	NewQuickPanelBtn(constants.CmdLater, tg.CmdTypeCallback, i18n.Emoji("Later"), "Later"),
	NewQuickPanelBtn(constants.CmdInlineQuerySearchEveryWhere, tg.CmdTypeInlineQueryCurrentChat, i18n.Emoji("Search"), "Search"),
	NewQuickPanelBtn(constants.CmdShowFiles, tg.CmdTypeCallback, i18n.Emoji("Files"), "Files"),
	NewQuickPanelBtn(constants.CmdShowChecklists, tg.CmdTypeCallback, i18n.Emoji("Checklists"), "Checklists"),
	NewQuickPanelBtn(constants.CmdShowPostpone, tg.CmdTypeCallback, i18n.Emoji("Postpone"), "Postpone"),
	NewQuickPanelBtn(constants.CmdShowReadChecklist, tg.CmdTypeCallback, i18n.Emoji("Read"), "Read"),
	NewQuickPanelBtn(constants.CmdShowWatchChecklist, tg.CmdTypeCallback, i18n.Emoji("Watch"), "Watch"),
	NewQuickPanelBtn(constants.CmdShowShopChecklist, tg.CmdTypeCallback, i18n.Emoji("Shop"), "Shop"),
	NewQuickPanelBtn(constants.CmdWebAppHabits, tg.CmdTypeWebApp, i18n.Emoji("Habits"), "Habits"),
}

var (
	QuickPanelAddButton = "➕"
	QuickPanelDelButton = "➖"
)

func NewQuickPanelBtn(cmd, cmdType, emoji, description string) QuickPanelBtn {
	return QuickPanelBtn{cmd, cmdType, emoji, description}
}

func (c *Config) AddPanelButton(button string) bool {
	// Does this button already exist?
	for _, curBtn := range c.raw.QuickPanelCommands {
		if curBtn == button {
			return false
		}
	}
	c.raw.QuickPanelCommands = append(c.raw.QuickPanelCommands, button)
	return true
}

func (c *Config) QuickPanelCmds() []string {
	return c.raw.QuickPanelCommands
}

func (c *Config) HasQuickPanelCmd(cmd string) bool {
	for _, pref := range c.raw.QuickPanelCommands {
		if cmd == pref {
			return true
		}
	}
	return false
}

func (c *Config) DelPanelButton(toDelete string) bool {
	var newButtons []string
	found := false // Was the target
	for _, curBtn := range c.raw.QuickPanelCommands {
		if curBtn == toDelete {
			found = true
		} else {
			newButtons = append(newButtons, curBtn)
		}
	}
	c.raw.QuickPanelCommands = newButtons
	return found
}
