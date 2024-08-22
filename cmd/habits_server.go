package main

import (
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/spf13/afero"

	"zakirullin/stuffbot/config"
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/habits"
	"zakirullin/stuffbot/internal/journal"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/txt"
)

// TODO graceful shutdown etc
func habitsServer() {
	router := http.NewServeMux()
	// TODO add hashing or secrets
	// TODO before release habits_v2 => habits
	router.HandleFunc("GET /habits_v2/{userID}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
		if err != nil {
			w.Write([]byte("can't parse userID"))
		}

		userPath := path.Join(config.Config.StoragePath, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, afero.NewOsFs())
		if err != nil {
			w.Write([]byte("can't init userFS"))
		}

		str, err := habits.Render(userID, userFS)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Write(str)
	})

	router.HandleFunc("POST /habits_v2/{userID}/{habitName}/{yearDay}/{status}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
		if err != nil {
			w.Write([]byte("can't parse userID"))
		}

		yearDay, err := strconv.ParseInt(r.PathValue("yearDay"), 10, 32)
		if err != nil {
			w.Write([]byte("can't parse yearDay"))
		}

		status, err := strconv.ParseInt(r.PathValue("status"), 10, 32)
		if err != nil {
			w.Write([]byte("can't parse status"))
		}

		habitName := r.PathValue("habitName")

		userPath := path.Join(config.Config.StoragePath, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, afero.NewOsFs())
		if err != nil {
			w.Write([]byte("can't init user fs"))
		}

		userHabits, err := habits.Habits(userFS, time.Now().Year())
		if err != nil {
			w.Write([]byte("can't read habits"))
		}

		if _, ok := userHabits[habitName]; !ok {
			userHabits[habitName] = make(habits.Year)
		}
		userHabits[habitName][int(yearDay)] = int(status)
		err = habits.Write(userFS, time.Now().Year(), userHabits)
		if err != nil {
			w.Write([]byte("can't write habits"))
		}

		var emoji string
		if habitName == habits.MoodHabit {
			if int(status) < len(habits.MoodEmojis) {
				emoji = habits.MoodEmojis[status]
			}
		} else {
			emoji, _ = userFS.Read(fs.DirHabits, fs.Filename(habitName))
			if emoji == "" {
				emoji = i18n.Emoji(habitName)
			}
			if emoji == "" {
				emoji = "⚡️"
			}
		}

		userConf := userconfig.NewConfig(userFS, userID, config.Config.ConfigFilename)
		err = journal.AddEmoji(userFS, emoji, userConf.Timezone())
		if err != nil {
			w.Write([]byte("can't write journal"))
		}
	})

	// TODO before release, don't panic if we don't want habits
	err := http.ListenAndServe(":81", router)
	if err != nil {
		panic(err)
	}
}
