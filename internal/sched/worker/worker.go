package worker

import (
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/exp/slog"

	"zakirullin/stuffbot/internal"
	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/sched"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/txt"
)

func MoveDueTasksToToday(
	storagePath,
	configFilename string,
	fsBackend afero.Fs,
	telegram internal.TGInterface,
) error {
	rootFS, err := fs.NewFS(storagePath, fsBackend)
	if err != nil {
		return fmt.Errorf("schedule worker: can't create FS: %s", err)
	}

	userDirs, err := rootFS.FilesAndDirs(fs.DirRoot)
	if err != nil {
		return fmt.Errorf("schedule worker: %w", err)
	}
	// TODO release
	// userDirs = fs.OnlyUserDirs(userDirs)

	for _, userDir := range userDirs {
		userID, err := strconv.ParseInt(userDir.Name, 10, 64)
		if err != nil {
			return fmt.Errorf("schedule worker: can't parse user ID: %s", err)
		}
		userPath := path.Join(storagePath, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, fsBackend)
		if err != nil {
			return fmt.Errorf("schedule worker: can't create FS: %s", err)
		}

		userconf := userconfig.NewConfig()
		userconfPath := userFS.UnsafePath("", configFilename)
		err = userconf.LoadOrCreate(userconfPath)
		if err != nil {
			return fmt.Errorf("schedule worker: can't load user config: %s", err)
		}

		for _, schedule := range userconf.Schedules() {
			if time.Now().Unix() < schedule.ScheduleAt {
				continue
			}
			err = moveTaskToToday(schedule.Filename, userFS)
			if err != nil {
				slog.Error("schedule worker: can't move", "err", err)
			}

			bot := internal.NewBot(userID, telegram, userFS, db.NewDB(), userconf)
			bot.ShowTodayTasks(nil)

			slog.Debug("Scheduled task moved to today", schedule.Filename, "filename")
			if len(schedule.Cron) != 0 {
				runAt := sched.Next(schedule.Cron)
				userconf.AddToSchedule(schedule.Filename, runAt, schedule.Cron)
				slog.Debug("Task was rescheduled", "filename", schedule.Filename, "schedule", schedule.Cron, "runAt", runAt)
				continue
			}

			err = userconf.LoadOrCreate(userconfPath)
			if err != nil {
				return fmt.Errorf("schedule worker: can't load user config before save: %s", err)
			}

			userconf.DelFromSchedule(schedule.Filename)
			err = userconf.Save(userconfPath)
			if err != nil {
				return fmt.Errorf("schedule worker: can't save user config: %s", err)
			}
		}
	}

	return nil
}

func moveTaskToToday(filename string, userFS *fs.FS) error {
	dirsToLookFor := []string{fs.DirLater, fs.DirArchive}
	for _, dir := range dirsToLookFor {
		filenames, err := userFS.FilesAndDirs(dir)
		if err != nil {
			return fmt.Errorf("moveTaskForToday: %w", err)
		}

		for _, f := range filenames {
			if f.Name == filename {
				err = userFS.Rename(dir, filename, fs.DirToday, filename)
				if err != nil {
					return fmt.Errorf("moveTaskForToday: can't rename: %w", err)
				}
			}
		}
	}

	return nil
}
