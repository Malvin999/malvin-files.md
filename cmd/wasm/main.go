package main

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"syscall/js"
	_ "time/tzdata" // for was env we need timezone database

	"zakirullin/stuffbot/internal"
	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/tg"
)

var (
	reply func(u internal.Update)
	chat  *tg.FakeTG
)

type Update struct {
	Message string
	Command *tg.Cmd
}

type Response struct {
	Messages []tg.Message
}

func Reply(_ js.Value, args []js.Value) interface{} {
	//callAsync("hi", func(result js.Value, err error) {
	//	if err != nil {
	//		sendToJS(fmt.Sprintf("Error: %v\n", err))
	//		return
	//	}
	//	sendToJS(result.String())
	//})
	upd := tg.NewUpd(-1, args[0].String())
	go reply(upd)
	//go readFile("file.md")

	return nil
}

func main() {
	fs.Exists = exists
	fs.ReadFile = readFile
	fs.WriteFile = writeFile
	fs.ReadDir = readDir
	initBot()
	js.Global().Set("reply", js.FuncOf(Reply))

	select {}
}

func callAsync(funcName string, callback func(js.Value, error), args ...any) {
	promise := js.Global().Call(funcName, args)

	var successFunc, errorFunc js.Func

	successFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer successFunc.Release()
		defer errorFunc.Release()
		callback(args[0], nil)
		return nil
	})

	errorFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer successFunc.Release()
		defer errorFunc.Release()
		callback(js.Undefined(), fmt.Errorf("error: %v", args[0]))
		return nil
	})

	promise.Call("then", successFunc).Call("catch", errorFunc)
}

func sendDueResponsesToJS() {
	var r Response
	r.Messages = chat.Messages
	if chat.EditedMessages != nil {
		r.Messages = append(r.Messages, chat.EditedMessages...)
	}

	chat.Messages = nil
	chat.EditedMessages = nil

	sendToJS(fmt.Sprintf("%v", r))
}

func sendToJS(vals ...any) {
	js.Global().Call("receive", vals...)
}

func initBot() {
	//opts := &tint.Options{
	//	Level: slog.LevelDebug,
	//}
	//logger := slog.New(tint.NewHandler(os.Stderr, opts))
	//slog.SetDefault(logger)

	// For GUI app we don't have required .env params
	//_ = godotenv.Load()
	//err := config.LoadGUIConfig()
	//if err != nil {
	//	panic(fmt.Sprintf("Error loading cfg: %s\n", err))
	//}

	// TODO move to embed
	//err = i18n.LoadLangFile("i18n/ru.json")
	//if err != nil {
	//	panic(fmt.Sprintf("Error loading i18n: %s\n", err))
	//}

	reply = func(u internal.Update) {
		defer func() {
			err := recover()
			if err != nil {
				debug.PrintStack()
				slog.Error("Bot panic", "err", err)
			}
		}()

		userID := u.UserID()

		userPath := ""
		userFS, err := fs.NewFS(userPath, NewJSFS())
		if err != nil {
			sendToJS(fmt.Sprintf("Bot error: can't create fs: %v", err))
		}
		err = userFS.CreateDirsIfNotExist()
		if err != nil {
			sendToJS(fmt.Sprintf("Bot error: can't create user dirs: %v", err))
		}

		confFilename := "config.json"
		userconf := userconfig.NewConfig(userFS, userID, confFilename)
		err = userconf.CreateDefaultIfNotExists()
		if err != nil {
			sendToJS("Bot error: can't create default user config")
		}

		if chat == nil {
			chat = tg.NewFakeTG()
		}
		bot := internal.NewBot(userID, chat, userFS, db.NewDB(userID), userconf)
		if err := bot.Reply(u); err != nil {
			sendToJS("Bot error", "err", err)
		}

		sendDueResponsesToJS()
	}
}

//func send(update Update) Response {
//	if update.Command != nil {
//		_ = reply(tg.NewUpdCmd(1, *update.Command))
//	} else {
//		_ = reply(tg.NewUpd(1, update.Message))
//	}
//
//	var r Response
//	r.Messages = chat.Messages
//	if chat.EditedMessages != nil {
//		r.Messages = append(r.Messages, chat.EditedMessages...)
//	}
//
//	chat.Messages = nil
//	chat.EditedMessages = nil
//
//	return r
//}
