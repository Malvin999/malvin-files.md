package journal

import (
	"fmt"
	"strings"
	"time"

	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/pkg/txt"
)

var now = time.Now

func AddRecord(botFs *fs.FS, noteFilename string) error {
	record, err := botFs.RestoreContent(fs.DirJournal, noteFilename)
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't get note content: %w", err)
	}

	journalFilename := now().Format("2024 January.md")
	exists, err := botFs.Exists(fs.DirJournal, journalFilename)
	if err != nil {
		return err
	}

	var md string
	if exists {
		md, err = botFs.Content(fs.DirJournal, journalFilename)
		if err != nil {
			return err
		}
		md = txt.NormNewLines(md)
		md = strings.TrimSpace(md)
	}

	header := fmt.Sprintf("#### %d, %s", now().Day(), now().Weekday())
	if !strings.Contains(md, header) {
		md = fmt.Sprintf("%s\n%s", md, header)
	}

	md = fmt.Sprintf("%s\n%s %s\n", md, now().Format("`13:01`"), record)

	return botFs.Put(fs.DirJournal, journalFilename, md)
}
