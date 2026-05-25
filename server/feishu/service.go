package feishu

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/channel"
	"github.com/larksuite/oapi-sdk-go/v3/channel/normalize"
	"github.com/larksuite/oapi-sdk-go/v3/channel/types"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"

	"github.com/zakirullin/files.md/server"
	"github.com/zakirullin/files.md/server/config"
	"github.com/zakirullin/files.md/server/db"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/userconfig"
)

type Service struct {
	cfg             Config
	channel         types.Channel
	eventDispatcher *dispatcher.EventDispatcher
	chat            *Chat
	mapper          identityMapper
	allowed         map[string]bool
	mu              sync.Mutex
	userChs         map[int64]chan server.Update
}

func Start(ctx context.Context, cfg Config) (*Service, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("feishu: app id and app secret are required")
	}

	client := lark.NewClient(cfg.AppID, cfg.AppSecret, lark.WithLogLevel(larkcore.LogLevelInfo))
	eventDispatcher := dispatcher.NewEventDispatcher("", "")
	wsClient := larkws.NewClient(
		cfg.AppID,
		cfg.AppSecret,
		larkws.WithEventHandler(eventDispatcher),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	ch := channel.NewChannel(client, wsClient)

	svc := &Service{
		cfg:             cfg,
		channel:         ch,
		eventDispatcher: eventDispatcher,
		chat:            NewChat(ch, client, cfg.EnableCardActions),
		mapper:          newIdentityMapper(cfg.DefaultUserID),
		allowed:         allowlist(cfg.AllowedOpenIDs),
		userChs:         make(map[int64]chan server.Update),
	}

	svc.installHandlers()

	go func() {
		if err := ch.Start(ctx); err != nil {
			slog.Error("Feishu channel stopped", "err", err)
		}
	}()

	return svc, nil
}

func (s *Service) installHandlers() {
	s.channel.OnReady(func() {
		slog.Info("Feishu channel ready")
	})
	s.channel.OnError(func(err error) {
		slog.Error("Feishu channel error", "err", err)
	})
	s.channel.OnReconnecting(func() {
		slog.Info("Feishu channel reconnecting")
	})
	s.channel.OnReconnected(func() {
		slog.Info("Feishu channel reconnected")
	})
	s.channel.OnDisconnected(func() {
		slog.Info("Feishu channel disconnected")
	})
	s.channel.OnReject(func(ctx context.Context, event *types.RejectEvent) error {
		slog.Info("Feishu message rejected", "messageID", event.MessageID, "senderID", event.SenderID, "reason", event.Reason)
		return nil
	})

	s.channel.OnMessage(func(ctx context.Context, msg *types.NormalizedMessage) error {
		if msg == nil || msg.UserID == "" {
			return nil
		}
		if !s.isAllowed(msg.UserID) {
			slog.Info("Feishu sender ignored", "openID", msg.UserID)
			return nil
		}

		userID := s.mapper.UserID(msg.UserID)
		slog.Info("Feishu message accepted", "openID", msg.UserID, "userID", userID, "chatID", msg.ChatID, "messageID", msg.MessageID)
		s.chat.RegisterUser(userID, msg.ChatID)
		s.chat.RegisterMessage(msg.MessageID)
		s.route(userID, NewMessageUpdate(userID, msg))
		return nil
	})

	s.eventDispatcher.OnP2CardActionTrigger(func(ctx context.Context, raw *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
		event := normalize.ParseCardAction(raw)
		if event == nil || event.Operator.OpenID == "" {
			slog.Info("Feishu card action ignored", "reason", "missing_operator")
			return cardToast("error", "无法识别操作"), nil
		}
		if !s.isAllowed(event.Operator.OpenID) {
			slog.Info("Feishu card action ignored", "openID", event.Operator.OpenID)
			return cardToast("error", "没有权限"), nil
		}

		userID := s.mapper.UserID(event.Operator.OpenID)
		slog.Info("Feishu card action accepted", "openID", event.Operator.OpenID, "userID", userID, "chatID", event.ChatID, "messageID", event.MessageID, "action", event.Action.Name)
		s.chat.RegisterUser(userID, event.ChatID)
		s.chat.RegisterMessage(event.MessageID)

		update := NewCardActionUpdate(userID, event)
		if update.Cmd() == nil {
			slog.Info("Feishu card action ignored", "openID", event.Operator.OpenID, "reason", "missing_command")
			return cardToast("error", "无法识别按钮"), nil
		}
		slog.Info("Feishu card command", "userID", userID, "command", update.Cmd().Name, "params", update.Cmd().Params)
		s.chat.SuppressNextHome()
		if err := s.processUpdate(userID, update); err != nil {
			slog.Error("Feishu card action failed", "userID", userID, "err", err)
			return cardToast("error", "处理失败"), nil
		}
		return cardToast("success", "已处理"), nil
	})
}

func (s *Service) isAllowed(openID string) bool {
	if len(s.allowed) == 0 {
		return true
	}
	return s.allowed[openID]
}

func (s *Service) route(userID int64, update server.Update) {
	s.mu.Lock()
	userCh, ok := s.userChs[userID]
	if !ok {
		userCh = make(chan server.Update, 100)
		s.userChs[userID] = userCh
		go s.supervisor(userID, userCh)
	}
	s.mu.Unlock()

	userCh <- update
}

func (s *Service) supervisor(userID int64, updates <-chan server.Update) {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("Feishu bot panic", "userID", userID, "err", err, "stacktrace", string(debug.Stack()))
				}
			}()
			s.processUserUpdates(userID, updates)
		}()
		time.Sleep(time.Second)
		slog.Info("Restarting Feishu worker", "userID", userID)
	}
}

func (s *Service) processUserUpdates(userID int64, updates <-chan server.Update) {
	for update := range updates {
		if err := s.processUpdate(userID, update); err != nil {
			slog.Error("Feishu bot error", "err", err)
		}
	}
}

func (s *Service) processUpdate(userID int64, update server.Update) error {
	bot, err := newBot(s.chat, userID)
	if err != nil {
		return fmt.Errorf("can't create bot: %w", err)
	}
	return bot.Reply(update)
}

func cardToast(toastType, content string) *callback.CardActionTriggerResponse {
	return &callback.CardActionTriggerResponse{
		Toast: &callback.Toast{
			Type:    toastType,
			Content: content,
		},
	}
}

func newBot(chat server.Chat, userID int64) (*server.Bot, error) {
	userFS, err := fs.NewUserFS(userID)
	if err != nil {
		return nil, fmt.Errorf("can't create fs: %w", err)
	}
	if err := userFS.CreateSystemDirs(); err != nil {
		return nil, fmt.Errorf("can't create user dirs: %w", err)
	}

	userconf := userconfig.NewConfig(userFS, userID, config.ServerCfg.ConfigFilename)
	if err := userconf.CreateDefaultIfNotExists(); err != nil {
		return nil, fmt.Errorf("can't create default user config: %w", err)
	}

	return server.NewBot(userID, chat, userFS, db.NewDB(userID), userconf), nil
}
