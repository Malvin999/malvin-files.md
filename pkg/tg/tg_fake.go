package tg

import (
	"io"
)

type FakeTG struct {
	SentTexts      []string
	LastSentText   string
	EditedText     string
	SentKeyboard   *Keyboard
	EditedKeyboard *Keyboard
}

func NewFakeTG() *FakeTG {
	return &FakeTG{}
}

func (f *FakeTG) Send(userID int64, text string, kb *Keyboard, markup string) (int, error) {
	f.LastSentText = text
	f.SentTexts = append(f.SentTexts, text)
	f.SentKeyboard = kb

	return -2, nil
}

func (f *FakeTG) Edit(userID int64, msgID int, text string, kb *Keyboard, markup string) error {
	f.EditedText = text
	f.EditedKeyboard = kb

	return nil
}

func (f *FakeTG) Del(userID int64, msgID int) error {
	return nil
}

func (f *FakeTG) AnswerCallbackQuery(queryID string, text string) error {
	return nil
}

func (f *FakeTG) AnswerInlineQuery(queryID string, results []interface{}, cacheTime int, offset string) error {
	return nil
}

func (f *FakeTG) DownloadFile(fileID string, writer io.Writer) (string, error) {
	return "", nil
}
