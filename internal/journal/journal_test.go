package journal

import (
	"testing"
	"time"

	"zakirullin/stuffbot/internal/fs"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_AddRecord(t *testing.T) {
	r := require.New(t)
	now = func() time.Time {
		return time.Date(2023, 05, 30, 10, 04, 36, 0, time.UTC)
	}

	type testcase struct {
		name                string
		md                  string
		note                string
		want                string
		journalHeaderFormat string
	}

	tests := []testcase{
		{
			name: "Empty MD",
			note: "note 1",
			want: "#### 30, Tuesday\n\nnote 1\n",
		},
		{
			name: "No Headers",
			md:   "some text",
			note: "note 1",
			want: "some text\n\n#### 30, Tuesday\n\nnote 1\n",
		},
		{
			name: "Bare header",
			md:   "#### 30, Tuesday\n",
			note: "note 1",
			want: "#### 30, Tuesday\n\nnote 1\n",
		},
		{
			name: "Bare headers",
			md:   "#### 30, Tuesday\n\n#### 31, Friday\n",
			note: "note 1",
			want: "#### 30, Tuesday\n\nnote 1\n\n#### 31, Friday\n",
		},
		{
			name: "New daily note",
			md:   "#### 29, Tuesday\n\nnote 1",
			note: "note 2",
			want: "#### 29, Tuesday\n\nnote 1\n\n#### 30, Tuesday\n\nnote 2\n",
		},
		{
			name: "Append daily note",
			md:   "#### 29, Tuesday\nnote 1\n\n#### 30, Tuesday\nnote 2",
			note: "note 3",
			want: "#### 29, Tuesday\n\nnote 1\n\n#### 30, Tuesday\n\nnote 2\n\n---\n\nnote 3\n",
		},
		{
			name: "Append daily note",
			md:   "#### 29, Tuesday\n\nnote 1\n\n#### 30, Tuesday\n\nnote 2\n",
			note: "note 3",
			want: "#### 29, Tuesday\n\nnote 1\n\n#### 30, Tuesday\n\nnote 2\n\n---\n\nnote 3\n",
		},
		{
			name: "Append daily note with custom header format",
			md:   "#### 29, Tuesday\n* note 1\n\n#### 30, Tuesday\nsome text\n* note 2",
			note: "note 3",
			want: "#### 29, Tuesday\n* note 1\n\n#### 30, Tuesday\n\nsome text\n* note 2\n\n#### 30.05.2023\n\nnote 3\n",
		},
		{
			name: "Higher Level Header",
			md:   "#### 30, Tuesday\n\nnote 1\n\n## Some Header\n\nnote 2\n",
			note: "note 3",
			want: "#### 30, Tuesday\n\nnote 1\n\n---\n\nnote 3\n\n## Some Header\n\nnote 2\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			botFS, err := fs.NewFS("/", afero.NewMemMapFs())
			r.NoError(err)
			botFS.Put(fs.DirJournal, "2023 May.md", test.md)
			botFS.Put(fs.DirToday, "note", test.note)

			AddRecord(botFS, "note.md")
			md, err := botFS.Content(fs.DirJournal, "2023 May.md")
			r.NoError(err)
			r.Equal(test.want, md)
		})
	}
}
