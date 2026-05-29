package feishu

import (
	"testing"

	"github.com/larksuite/oapi-sdk-go/v3/channel/types"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/stretchr/testify/require"

	"github.com/zakirullin/files.md/server/pkg/tg"
)

func TestCommandFromTextOnlyAcceptsSingleSlashCommand(t *testing.T) {
	cmd := commandFromText("/settings")
	require.NotNil(t, cmd)
	require.Equal(t, "settings", cmd.Name)

	require.Nil(t, commandFromText("/later buy milk"))
	require.Nil(t, commandFromText("buy milk jj"))
}

func TestCommandFromCardAction(t *testing.T) {
	event := &types.CardActionEvent{
		Action: types.CardActionPayload{
			Value: map[string]interface{}{
				cardCommandKey: map[string]interface{}{
					"n": "mv_later",
					"p": []interface{}{"abc123"},
					"t": "cmd",
				},
			},
		},
	}

	cmd := commandFromCardAction(event)
	require.NotNil(t, cmd)
	require.Equal(t, "mv_later", cmd.Name)
	require.Equal(t, []string{"abc123"}, cmd.Params)
}

func TestIdentityMapperUsesConfiguredUserID(t *testing.T) {
	mapper := newIdentityMapper(10001)

	require.Equal(t, int64(10001), mapper.UserID("ou_1"))
	require.Equal(t, int64(10001), mapper.UserID("ou_2"))
}

func TestMediaIDRoundTrip(t *testing.T) {
	id := mediaID("image", "img_xxx", "photo.webp", "om_xxx")

	mediaType, fileKey, ext, messageID, ok := parseMediaID(id)
	require.True(t, ok)
	require.Equal(t, "image", mediaType)
	require.Equal(t, "img_xxx", fileKey)
	require.Equal(t, ".webp", ext)
	require.Equal(t, "om_xxx", messageID)
}

func TestPureImageMessageDoesNotUseImageKeyAsCaption(t *testing.T) {
	update := NewMessageUpdate(10001, &types.NormalizedMessage{
		MessageID:      "om_xxx",
		Content:        "![image](img_xxx)",
		RawContentType: "image",
		Resources: []types.Resource{{
			Type:    "image",
			FileKey: "img_xxx",
		}},
	})

	require.NotEmpty(t, update.imageID)
	require.Empty(t, update.Caption())
}

func TestPostMessageUsesTextAsImageCaption(t *testing.T) {
	rawContent := `{"title":"","content":[[{"tag":"text","text":"早上整理了一下思路"},{"tag":"img","image_key":"img_v2_xxx"}]]}`
	update := NewMessageUpdate(10001, &types.NormalizedMessage{
		MessageID:      "om_xxx",
		Content:        "[rich text message]",
		RawContentType: "post",
		RawEvent:       rawPostEvent(rawContent),
	})

	require.Equal(t, "早上整理了一下思路", update.Caption())
	require.NotContains(t, update.Caption(), "img_v2_xxx")
	require.Equal(t, "早上整理了一下思路", update.MsgText())
	require.NotEmpty(t, update.imageID)

	mediaType, fileKey, ext, messageID, ok := parseMediaID(update.imageID)
	require.True(t, ok)
	require.Equal(t, "image", mediaType)
	require.Equal(t, "img_v2_xxx", fileKey)
	require.Equal(t, ".png", ext)
	require.Equal(t, "om_xxx", messageID)
}

func TestPostTextMessageSupportsLocaleWrapper(t *testing.T) {
	rawContent := `{"zh_cn":{"title":"日记","content":[[{"tag":"text","text":"只是一段文字"}]]}}`
	update := NewMessageUpdate(10001, &types.NormalizedMessage{
		MessageID:      "om_xxx",
		Content:        "[rich text message]",
		RawContentType: "post",
		RawEvent:       rawPostEvent(rawContent),
	})

	require.Empty(t, update.Caption())
	require.Empty(t, update.imageID)
	require.Equal(t, "**日记**\n\n只是一段文字", update.MsgText())
}

func TestCardCommandJSONRoundTrip(t *testing.T) {
	card, err := keyboardCard("Saved!", tg.NewKeyboard([]tg.Row{
		tg.NewBtn("Later", tg.NewCmd("mv_later", []string{"abc123"})),
	}))
	require.NoError(t, err)
	require.Contains(t, card, cardCommandKey)
	require.Contains(t, card, "mv_later")
	require.Contains(t, card, "已记录")
	require.Contains(t, card, "稍后")
}

func TestKeyboardCardSkipsSeparators(t *testing.T) {
	card, err := keyboardCard("Settings:", tg.NewKeyboard([]tg.Row{
		tg.NewBtn("-", tg.NewCmd("nothing", nil)),
		tg.NewBtn("🏠 Home", tg.NewCmd("home", nil)),
	}))
	require.NoError(t, err)
	require.NotContains(t, card, "\"content\":\"-\"")
	require.Contains(t, card, "首页")
}

func TestKeyboardCardSkipsUnsupportedCommandTypes(t *testing.T) {
	card, err := keyboardCard("Settings:", tg.NewKeyboard([]tg.Row{
		tg.NewBtn("Search", tg.NewCustomCmd("search", nil, tg.CmdTypeInlineQueryCurrentChat)),
		tg.NewBtn("Later", tg.NewCmd("later", nil)),
	}))
	require.NoError(t, err)
	require.NotContains(t, card, "search")
	require.Contains(t, card, "later")
	require.Contains(t, card, "稍后")
}

func TestIsHomeCard(t *testing.T) {
	require.True(t, isHomeCard("3 items", tg.NewKeyboard([]tg.Row{
		tg.NewRow(
			tg.NewBtn("Item", tg.NewCmd("c", []string{"abc"})),
			tg.NewBtn("Move", tg.NewCmd("s_move", []string{"abc"})),
		),
		tg.NewBtn("🏠 Home", tg.NewCmd("home", nil)),
	})))

	require.False(t, isHomeCard("Saved!", tg.NewKeyboard([]tg.Row{
		tg.NewBtn("Later", tg.NewCmd("mv_later", []string{"abc"})),
	})))
}

func rawPostEvent(content string) *larkim.P2MessageReceiveV1 {
	return &larkim.P2MessageReceiveV1{
		Event: &larkim.P2MessageReceiveV1Data{
			Message: &larkim.EventMessage{
				Content: &content,
			},
		},
	}
}
