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
	NewQuickPanelBtn(constants.InlineQuerySearchEveryWhere, tg.CmdTypeInlineQueryCurrentChat, i18n.Emoji("Search"), "Search"),
	NewQuickPanelBtn(constants.CmdShowDoc, tg.CmdTypeCallback, i18n.EmDocs, "Documents"),
	NewQuickPanelBtn(constants.CmdShowChecklists, tg.CmdTypeCallback, i18n.EmCheckList, "Checklists"),
	NewQuickPanelBtn(constants.CmdShowPostpone, tg.CmdTypeCallback, i18n.EmPostpone, "Postpone"),
	NewQuickPanelBtn(constants.CmdShowReadChecklist, tg.CmdTypeCallback, i18n.Emoji("Read"), "Read"),
	NewQuickPanelBtn(constants.CmdShowWatchChecklist, tg.CmdTypeCallback, i18n.Emoji("Watch"), "Watch"),
	NewQuickPanelBtn(constants.CmdShowShopChecklist, tg.CmdTypeCallback, i18n.Emoji("Shop"), "Shop"),
}

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
