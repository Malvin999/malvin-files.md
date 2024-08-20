package userconfig

import (
	"time"
)

type FakeConfig struct {
}

func NewFakeConfig() *FakeConfig {
	return &FakeConfig{}
}

func (c *FakeConfig) CreateDefaultIfNotExists() error {
	return nil
}

func (c *FakeConfig) SetPomodoroDuration(duration time.Duration) error {
	return nil
}

func (c *FakeConfig) PomodoroDuration() time.Duration {
	return time.Duration(defaultConfig.PomodoroDurationInMinutes * int64(time.Minute))
}

func (c *FakeConfig) Schedules() ([]Schedule, error) {
	return nil, nil
}

func (c *FakeConfig) AddToSchedule(filename string, scheduleAt int64, cron string) error {
	return nil
}

func (c *FakeConfig) DelFromSchedule(filename string, scheduledAt int64) error {
	return nil
}

func (c *FakeConfig) ShouldSplitChecklist(checklist string) bool {
	return false
}

func (c *FakeConfig) AddQuickCmd(cmd string) error {
	return nil
}

func (c *FakeConfig) QuickCmds() ([]string, error) {
	return nil, nil
}

func (c *FakeConfig) DelQuickCmd(cmd string) error {
	return nil
}

func (c *FakeConfig) AddMoveToCmd(cmd string) error {
	return nil
}

func (c *FakeConfig) MoveToCmds() ([]string, error) {
	return nil, nil
}

func (c *FakeConfig) DelMoveToCmd(cmd string) error {
	return nil
}
