package habits

// TODO one known bug - it won't correctly work
// if our week falls into 2 different years

import (
	_ "embed"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"zakirullin/stuffbot/internal/fs"
)

//go:embed testdata/month_habits.md
var monthMD string

//go:embed testdata/last_month_habits.md
var lastMonthMD string

//go:embed testdata/two_months_habits.md
var twoMonthsMD string

func TestHabits(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write(fs.DirInsights, "1970 Habits.md", monthMD)

	habits, err := Habits(userFS, 1970)
	r.NoError(err)

	r.Len(habits, 5)
	year, ok := habits["Went to gym"]
	r.True(ok)

	r.Len(year, 31)

	r.EqualValues(Year{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 1, 7: 0, 8: 0, 9: 0, 10: 0, 11: 1, 12: 0, 13: 0, 14: 1, 15: 0, 16: 0, 17: 0, 18: 1, 19: 0, 20: 1, 21: 0, 22: 0, 23: 1, 24: 0, 25: 1, 26: 0, 27: 1, 28: 0, 29: 1, 30: 0, 31: 1}, year)
}

func TestLastMonthHabits(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write(fs.DirInsights, "1970 Habits.md", lastMonthMD)

	habits, err := Habits(userFS, 1970)
	r.NoError(err)

	r.Len(habits, 1)
	year, ok := habits["Habit"]
	r.True(ok)

	r.Len(year, 31)

	completed, ok := year[335]
	r.True(ok)
	r.Equal(0, completed)

	completed, ok = year[365]
	r.True(ok)
	r.Equal(1, completed)
}

func TestLastWeekHabitsWhenWeekFallsIntoTwoMonths(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write(fs.DirInsights, "1970 Habits.md", twoMonthsMD)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(1970, time.September, 30, 0, 0, 0, 0, time.Local)
	}

	habits, err := LastWeekHabits(userFS)
	r.NoError(err)
	r.Len(habits, 1)
	r.Len(habits["Habit"], 7)
	r.EqualValues(Year{271: 0, 272: 1, 273: 0, 274: 0, 275: 0, 276: 1, 277: 0}, habits["Habit"])
}

func TestLastMonthHabitsMoods(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write(fs.DirInsights, "1970 Habits.md", monthMD)

	habits, err := Habits(userFS, 1970)
	r.NoError(err)

	year, ok := habits["Mood"]
	r.True(ok)

	r.Len(year, 31)

	r.EqualValues(Year{1: 5, 2: 0, 3: 3, 4: 1, 5: 0, 6: 5, 7: 5, 8: 0, 9: 0, 10: 0, 11: 5, 12: 0, 13: 5, 14: 2, 15: 4, 16: 1, 17: 0, 18: 5, 19: 0, 20: 4, 21: 0, 22: 5, 23: 0, 24: 5, 25: 4, 26: 0, 27: 5, 28: 4, 29: 0, 30: 5, 31: 0}, year)
}

func TestWrite(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	userFS.Write("insights", "2024 Habits.md", monthMD)

	habits, err := Habits(userFS, 2024)
	r.NoError(err)

	err = Write(userFS, 2024, habits)
	r.NoError(err)

	updatedMonthMD, err := userFS.Read("insights", "2024 Habits.md")
	r.NoError(err)

	r.Equal(monthMD, updatedMonthMD)
}