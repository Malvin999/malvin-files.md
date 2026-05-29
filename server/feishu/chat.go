package feishu

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/channel/types"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/zakirullin/files.md/server/pkg/tg"
)

type Chat struct {
	channel           types.Channel
	client            *lark.Client
	enableCardActions bool

	mu         sync.RWMutex
	chatIDs    map[int64]string
	messageIDs map[int]string
	msgSeq     atomic.Int64
	suppress   int32
}

func NewChat(ch types.Channel, client *lark.Client, enableCardActions bool) *Chat {
	return &Chat{
		channel:           ch,
		client:            client,
		enableCardActions: enableCardActions,
		chatIDs:           make(map[int64]string),
		messageIDs:        make(map[int]string),
	}
}

func (c *Chat) RegisterUser(userID int64, chatID string) {
	if chatID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.chatIDs[userID] = chatID
}

func (c *Chat) RegisterMessage(messageID string) {
	if messageID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageIDs[stableIntID(messageID)] = messageID
}

func (c *Chat) Send(userID int64, text string, kb *tg.Keyboard, markup string) (int, error) {
	if c.shouldSuppress(text, kb) {
		return int(c.msgSeq.Add(1)), nil
	}

	chatID, ok := c.chatID(userID)
	if !ok {
		return 0, fmt.Errorf("feishu send: no chat id for user %d", userID)
	}

	input := &types.SendInput{ChatID: chatID}
	if kb != nil && c.enableCardActions {
		card, err := keyboardCard(plainText(text), kb)
		if err != nil {
			return 0, fmt.Errorf("feishu send card: %w", err)
		}
		input.Card = card
	} else {
		input.Text = plainText(text)
	}

	resp, err := c.channel.Send(context.Background(), input)
	if err != nil {
		return 0, fmt.Errorf("feishu send: %w", err)
	}

	if resp != nil && resp.MessageID != "" {
		return stableIntID(resp.MessageID), nil
	}
	return int(c.msgSeq.Add(1)), nil
}

func (c *Chat) SendImages(userID int64, images []string) ([]int, error) {
	var ids []int
	for _, image := range images {
		id, err := c.Send(userID, image, nil, tg.MarkupHTML)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *Chat) SendReaction(userID int64, msgID int, reaction string) error {
	chatID, ok := c.chatID(userID)
	if !ok {
		return fmt.Errorf("feishu reaction: no chat id for user %d", userID)
	}

	input := &types.SendInput{
		ChatID:         chatID,
		Text:           "已记录",
		ReplyMessageID: c.messageID(msgID),
	}
	_, err := c.channel.Send(context.Background(), input)
	if err != nil {
		return fmt.Errorf("feishu reaction: %w", err)
	}
	return nil
}

func (c *Chat) Edit(userID int64, msgID int, text string, kb *tg.Keyboard, markup string) error {
	_, err := c.Send(userID, text, kb, markup)
	return err
}

func (c *Chat) Del(userID int64, msgID int) error {
	return nil
}

func (c *Chat) AnswerCallbackQuery(queryID string, text string) error {
	return nil
}

func (c *Chat) AnswerInlineQuery(queryID string, results []interface{}, cacheTime int, offset string) error {
	return nil
}

func (c *Chat) SuppressNextHome() {
	c.suppress = 1
}

func (c *Chat) DownloadFile(fileID string, outFile io.Writer) (string, error) {
	mediaType, fileKey, ext, messageID, ok := parseMediaID(fileID)
	if !ok {
		return "", fmt.Errorf("feishu download: invalid media id")
	}

	if messageID != "" && c.client != nil {
		return c.downloadMessageResource(messageID, fileKey, mediaType, ext, outFile)
	}

	bytes, err := c.channel.DownloadFile(context.Background(), fileKey, mediaType)
	if err != nil {
		return "", fmt.Errorf("feishu download: %w", err)
	}

	if _, err := io.Copy(outFile, bytesReader(bytes)); err != nil {
		return "", fmt.Errorf("feishu download write: %w", err)
	}
	return ext, nil
}

func (c *Chat) downloadMessageResource(messageID, fileKey, mediaType, ext string, outFile io.Writer) (string, error) {
	resourceType := "file"
	if mediaType == "image" {
		resourceType = "image"
	}

	req := larkim.NewGetMessageResourceReqBuilder().
		MessageId(messageID).
		FileKey(fileKey).
		Type(resourceType).
		Build()
	resp, err := c.client.Im.V1.MessageResource.Get(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("feishu download message resource: %w", err)
	}
	if !resp.Success() {
		return "", fmt.Errorf("feishu download message resource API error: %d - %s", resp.Code, resp.Msg)
	}

	if _, err := io.Copy(outFile, resp.File); err != nil {
		return "", fmt.Errorf("feishu download message resource write: %w", err)
	}
	return messageResourceExtension(mediaType, ext, resp.FileName, resp.Header.Get("Content-Type")), nil
}

func messageResourceExtension(mediaType, requestedExt, filename, contentType string) string {
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
		return ext
	}

	if mediaType == "image" {
		if ext := imageContentTypeExtension(contentType); ext != "" {
			return ext
		}
		if requestedExt != "" {
			return requestedExt
		}
		return ".png"
	}

	return requestedExt
}

func imageContentTypeExtension(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	case "image/heic":
		return ".heic"
	case "image/heif":
		return ".heif"
	default:
		return ""
	}
}

func (c *Chat) chatID(userID int64) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	chatID, ok := c.chatIDs[userID]
	return chatID, ok
}

func (c *Chat) messageID(stableID int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messageIDs[stableID]
}

func (c *Chat) shouldSuppress(text string, kb *tg.Keyboard) bool {
	if c.suppress == 0 {
		return false
	}
	if !isHomeCard(text, kb) {
		return false
	}
	c.suppress = 0
	return true
}

func isHomeCard(text string, kb *tg.Keyboard) bool {
	if kb == nil {
		return strings.TrimSpace(plainText(text)) == "What's on your mind?"
	}
	if len(kb.Btns) == 0 {
		return false
	}

	for _, row := range kb.Btns {
		switch typed := row.(type) {
		case tg.Btn:
			if typed.Cmd.Name != "home" && typed.Cmd.Name != "c" && typed.Cmd.Name != "s_move" {
				return false
			}
		case []tg.Btn:
			for _, btn := range typed {
				if btn.Cmd.Name != "home" && btn.Cmd.Name != "c" && btn.Cmd.Name != "s_move" {
					return false
				}
			}
		default:
			return false
		}
	}
	return true
}

func plainText(s string) string {
	s = regexp.MustCompile(`</?[^>]+>`).ReplaceAllString(s, "")
	return html.UnescapeString(s)
}

func bytesReader(b []byte) io.Reader {
	return bytes.NewReader(b)
}
