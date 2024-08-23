package tg

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type FakeUpd struct {
	userID           int64
	cmd              Cmd
	Msg              string
	PhotoID          string
	PhotoCaption     string
	ReplyToMessageID int
}

func NewFakeUpd(userID int64, msg string) *FakeUpd {
	return &FakeUpd{userID: userID, Msg: msg, ReplyToMessageID: -1}
}

func NewFakeUpdCmd(id int64, cmd Cmd) *FakeUpd {
	return &FakeUpd{userID: id, cmd: cmd}
}

func (m *FakeUpd) MsgText() string {
	return m.Msg
}

func (m *FakeUpd) UserID() int64 {
	return m.userID
}

func (m *FakeUpd) Cmd() *Cmd {
	if m.cmd.Name == "" {
		return nil
	}

	return &m.cmd
}

func (m *FakeUpd) MsgEntities() []tgbotapi.MessageEntity {
	return nil
}

func (m *FakeUpd) CaptionEntities() []tgbotapi.MessageEntity {
	return nil
}

func (m *FakeUpd) CallbackQueryID() (string, bool) {
	return "", true
}

func (m *FakeUpd) InlineQueryID() (string, bool) {
	return "", false
}

func (m *FakeUpd) InlineQuery() (string, bool) {
	return "", false
}

func (m *FakeUpd) InlineQueryOffset() int {
	return 0
}

func (m *FakeUpd) IsForwarded() bool {
	return false
}

func (m *FakeUpd) IsSentViaBot() bool {
	return false
}

func (m *FakeUpd) ReplyToMsgID() (int, bool) {
	return m.ReplyToMessageID, m.ReplyToMessageID != -1
}

func (m *FakeUpd) PhotoOrImageID() (string, bool) {
	if m.PhotoID != "" {
		return m.PhotoID, true
	}

	return "", false
}

func (m *FakeUpd) Caption() string {
	return m.PhotoCaption
}
