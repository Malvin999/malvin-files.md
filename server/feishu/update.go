package feishu

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/larksuite/oapi-sdk-go/v3/channel/types"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/zakirullin/files.md/server/pkg/tg"
)

const cardCommandKey = "files_md_cmd"

type Update struct {
	userID       int64
	text         string
	messageID    string
	createTimeMs int64
	replyToID    int
	cmd          *tg.Cmd
	callbackID   string
	imageID      string
	caption      string
}

func NewMessageUpdate(userID int64, msg *types.NormalizedMessage) *Update {
	text := strings.TrimSpace(msg.Content)
	resources := msg.Resources
	if msg.RawContentType == "post" {
		if postText, postResources, ok := parsePostMessage(msg); ok {
			text = postText
			resources = postResources
		}
	}

	u := &Update{
		userID:       userID,
		text:         text,
		messageID:    msg.MessageID,
		createTimeMs: msg.CreateTimeMs,
	}

	for _, resource := range resources {
		if resource.Type == "image" && resource.FileKey != "" {
			u.imageID = mediaID(resource.Type, resource.FileKey, resource.FileName, msg.MessageID)
			break
		}
	}

	if u.imageID != "" && msg.RawContentType != "image" {
		u.caption = u.text
	}

	if cmd := commandFromText(u.text); cmd != nil {
		u.cmd = cmd
	}

	return u
}

func parsePostMessage(msg *types.NormalizedMessage) (string, []types.Resource, bool) {
	rawContent := rawMessageContent(msg.RawEvent)
	if rawContent == "" {
		return "", nil, false
	}
	return parsePostContent(rawContent)
}

func rawMessageContent(rawEvent interface{}) string {
	event, ok := rawEvent.(*larkim.P2MessageReceiveV1)
	if !ok || event.Event == nil || event.Event.Message == nil || event.Event.Message.Content == nil {
		return ""
	}
	return *event.Event.Message.Content
}

func parsePostContent(rawContent string) (string, []types.Resource, bool) {
	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawContent), &contentMap); err != nil {
		return "", nil, false
	}

	bodyMap := postBody(contentMap)
	if bodyMap == nil {
		return "", nil, false
	}

	lines := make([]string, 0)
	if title, _ := bodyMap["title"].(string); strings.TrimSpace(title) != "" {
		lines = append(lines, fmt.Sprintf("**%s**", strings.TrimSpace(title)), "")
	}

	var resources []types.Resource
	if contentList, ok := bodyMap["content"].([]interface{}); ok {
		for _, paragraphInterface := range contentList {
			paragraph, ok := paragraphInterface.([]interface{})
			if !ok {
				continue
			}

			var lineParts []string
			for _, elInterface := range paragraph {
				el, ok := elInterface.(map[string]interface{})
				if !ok {
					continue
				}

				text, newResources := postElementText(el)
				if text != "" {
					lineParts = append(lineParts, text)
				}
				resources = append(resources, newResources...)
			}

			if line := strings.TrimSpace(strings.Join(lineParts, "")); line != "" {
				lines = append(lines, line)
			}
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), resources, len(lines) > 0 || len(resources) > 0
}

func postBody(contentMap map[string]interface{}) map[string]interface{} {
	if _, ok := contentMap["content"].([]interface{}); ok {
		return contentMap
	}

	for _, locale := range []string{"zh_cn", "en_us", "ja_jp", "zh_hk"} {
		if body, ok := contentMap[locale].(map[string]interface{}); ok {
			if _, hasContent := body["content"].([]interface{}); hasContent {
				return body
			}
		}
	}

	for _, v := range contentMap {
		body, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if _, hasContent := body["content"].([]interface{}); hasContent {
			return body
		}
	}

	return nil
}

func postElementText(el map[string]interface{}) (string, []types.Resource) {
	tag, _ := el["tag"].(string)
	text, _ := el["text"].(string)

	switch tag {
	case "text":
		return text, nil
	case "a":
		href, _ := el["href"].(string)
		label := text
		if label == "" {
			label = href
		}
		if href == "" {
			return label, nil
		}
		return fmt.Sprintf("[%s](%s)", label, href), nil
	case "at":
		userID, _ := el["user_id"].(string)
		userName, _ := el["user_name"].(string)
		if userID == "all" || userID == "all_members" {
			return "@all", nil
		}
		if userName != "" {
			return "@" + userName, nil
		}
		if userID != "" {
			return "@" + userID, nil
		}
		return "", nil
	case "img":
		imageKey, _ := el["image_key"].(string)
		if imageKey == "" {
			return "", nil
		}
		return "", []types.Resource{{Type: "image", FileKey: imageKey}}
	case "media":
		fileKey, _ := el["file_key"].(string)
		if fileKey == "" {
			return "", nil
		}
		return fmt.Sprintf(`<file key="%s"/>`, fileKey), []types.Resource{{Type: "file", FileKey: fileKey}}
	case "code_block":
		lang, _ := el["language"].(string)
		return fmt.Sprintf("\n```%s\n%s\n```\n", lang, text), nil
	case "hr":
		return "\n---\n", nil
	default:
		return text, nil
	}
}

func NewCardActionUpdate(userID int64, event *types.CardActionEvent) *Update {
	return &Update{
		userID:     userID,
		messageID:  event.MessageID,
		callbackID: event.EventID,
		cmd:        commandFromCardAction(event),
	}
}

func (u *Update) MsgText() string {
	return u.text
}

func (u *Update) UserID() int64 {
	return u.userID
}

func (u *Update) Cmd() *tg.Cmd {
	return u.cmd
}

func (u *Update) MsgEntities() []tgbotapi.MessageEntity {
	return nil
}

func (u *Update) CaptionEntities() []tgbotapi.MessageEntity {
	return nil
}

func (u *Update) CallbackQueryID() (string, bool) {
	return u.callbackID, u.callbackID != ""
}

func (u *Update) InlineQueryID() (string, bool) {
	return "", false
}

func (u *Update) InlineQuery() (string, bool) {
	return "", false
}

func (u *Update) InlineQueryOffset() int {
	return 0
}

func (u *Update) IsSentViaBot() bool {
	return false
}

func (u *Update) ReplyToMsgID() (int, bool) {
	return u.replyToID, u.replyToID != 0
}

func (u *Update) PhotoOrImageID() (string, bool) {
	return u.imageID, u.imageID != ""
}

func (u *Update) Caption() string {
	return u.caption
}

func (u *Update) MsgID() (int, bool) {
	return stableIntID(u.messageID), u.messageID != ""
}

func (u *Update) Time() (int, bool) {
	if u.createTimeMs <= 0 {
		return 0, false
	}
	return int(u.createTimeMs / 1000), true
}

func (u *Update) ChannelID() (int64, bool) {
	return 0, false
}

func (u *Update) ChannelName() (string, bool) {
	return "", false
}

func commandFromText(text string) *tg.Cmd {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return nil
	}

	fields := strings.Fields(trimmed)
	if len(fields) != 1 {
		return nil
	}

	name := strings.TrimPrefix(fields[0], "/")
	if name == "" {
		return nil
	}

	cmd := tg.NewCmd(name, nil)
	return &cmd
}

func commandFromCardAction(event *types.CardActionEvent) *tg.Cmd {
	if event == nil || event.Action.Value == nil {
		return nil
	}

	raw, ok := event.Action.Value[cardCommandKey]
	if !ok {
		return nil
	}

	var cmd tg.Cmd
	switch typed := raw.(type) {
	case string:
		if err := json.Unmarshal([]byte(typed), &cmd); err != nil {
			return nil
		}
	default:
		b, err := json.Marshal(typed)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(b, &cmd); err != nil {
			return nil
		}
	}

	if cmd.Name == "" {
		return nil
	}
	return &cmd
}

func stableIntID(s string) int {
	if s == "" {
		return 0
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum64() & uint64(math.MaxInt32))
}

func mediaID(mediaType, fileKey, filename, messageID string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" && mediaType == "image" {
		ext = ".png"
	}

	payload := map[string]string{
		"type": mediaType,
		"key":  fileKey,
		"ext":  ext,
		"msg":  messageID,
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func parseMediaID(id string) (mediaType, fileKey, ext, messageID string, ok bool) {
	var payload map[string]string
	if err := json.Unmarshal([]byte(id), &payload); err != nil {
		return "", "", "", "", false
	}

	mediaType = payload["type"]
	fileKey = payload["key"]
	ext = payload["ext"]
	messageID = payload["msg"]
	return mediaType, fileKey, ext, messageID, mediaType != "" && fileKey != ""
}
