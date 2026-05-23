package feishu

import (
	"encoding/json"
	"strings"

	"github.com/zakirullin/files.md/server/pkg/tg"
)

func keyboardCard(text string, kb *tg.Keyboard) (string, error) {
	elements := []map[string]interface{}{
		{
			"tag":     "markdown",
			"content": feishuText(text),
		},
	}

	if kb != nil {
		for _, row := range kb.Btns {
			columns := buttonsToColumns(row)
			if len(columns) == 0 {
				continue
			}
			elements = append(elements, map[string]interface{}{
				"tag":     "column_set",
				"columns": columns,
			})
		}
	}

	card := map[string]interface{}{
		"schema": "2.0",
		"config": map[string]interface{}{
			"update_multi": true,
		},
		"body": map[string]interface{}{
			"direction": "vertical",
			"elements":  elements,
		},
	}

	b, err := json.Marshal(card)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func buttonsToColumns(row tg.Row) []map[string]interface{} {
	switch typed := row.(type) {
	case tg.Btn:
		column, ok := buttonColumn(typed)
		if !ok {
			return nil
		}
		return []map[string]interface{}{column}
	case []tg.Btn:
		columns := make([]map[string]interface{}, 0, len(typed))
		for _, btn := range typed {
			column, ok := buttonColumn(btn)
			if !ok {
				continue
			}
			columns = append(columns, column)
		}
		return columns
	default:
		return nil
	}
}

func buttonColumn(btn tg.Btn) (map[string]interface{}, bool) {
	if btn.Name == "-" || btn.Cmd.Name == "nothing" {
		return nil, false
	}
	if btn.Cmd.Type != "" && btn.Cmd.Type != tg.CmdTypeCallback {
		return nil, false
	}

	return map[string]interface{}{
		"tag":    "column",
		"width":  "weighted",
		"weight": 1,
		"elements": []map[string]interface{}{
			{
				"tag": "button",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": feishuButtonLabel(btn.Name),
				},
				"type": "default",
				"behaviors": []map[string]interface{}{
					{
						"type": "callback",
						"value": map[string]interface{}{
							cardCommandKey: btn.Cmd,
						},
					},
				},
			},
		},
	}, true
}

func feishuText(text string) string {
	text = strings.TrimSpace(text)
	switch text {
	case "Saved!":
		return "已记录。现在可以选择去向，也可以先不处理。"
	case "Settings:":
		return "设置"
	case "What's on your mind?":
		return "想记录什么？"
	default:
		return text
	}
}

func feishuButtonLabel(label string) string {
	label = strings.TrimSpace(label)
	switch {
	case label == "👌":
		return "留在 Chat"
	case label == "Later":
		return "稍后"
	case strings.Contains(label, "To later"):
		return "稍后"
	case strings.Contains(label, "To tmrw"):
		return "明天"
	case strings.Contains(label, "To a day"):
		return "指定日期"
	case strings.Contains(label, "To File"):
		return "文件"
	case strings.Contains(label, "To Journal"):
		return "日记"
	case strings.Contains(label, "To Read"):
		return "阅读"
	case strings.Contains(label, "To Watch"):
		return "观看"
	case strings.Contains(label, "To Shop"):
		return "购买"
	case strings.Contains(label, "To Checklist"):
		return "清单"
	case strings.Contains(label, "Full mode"):
		return "完整模式"
	case strings.Contains(label, "Notes mode"):
		return "笔记模式"
	case strings.Contains(label, "Tasks mode"):
		return "任务模式"
	case strings.Contains(label, "Quick buttons"):
		return "快捷按钮"
	case strings.Contains(label, "Move to buttons"):
		return "分类按钮"
	case strings.Contains(label, "Timezone"):
		return "时区"
	case strings.Contains(label, "Home"):
		return "首页"
	default:
		return label
	}
}
