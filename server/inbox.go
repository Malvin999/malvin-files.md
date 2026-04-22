package server

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/pkg/txt"
)

var (
	Now       = time.Now
	mu        sync.Mutex
	userLocks map[string]*sync.Mutex

	inboxMarkerPrefix = regexp.MustCompile(`^- \[[ xX]\] `)
	inboxHeaderRegex  = regexp.MustCompile(`^#### `)
	inboxEntryPrefix  = regexp.MustCompile(`^(?:- \[[ xX]\] )?` + "`" + `\d{2}:\d{2}` + "`" + ` `)
)

// stripInboxEntryPrefix removes the optional `- [ ]` / `- [x] ` task marker
// and the `` `HH:MM` `` timestamp so only the entry body remains.
func stripInboxEntryPrefix(block string) string {
	return inboxEntryPrefix.ReplaceAllString(block, "")
}

// inboxBlockHash returns a stable identifier for an inbox block. The optional
// `- [ ]` / `- [x] ` task-marker prefix is stripped before hashing so the hash
// is the same regardless of completion state (a completed entry keeps the
// identity of the open entry it was toggled from).
func inboxBlockHash(block string) string {
	return fs.Hash(inboxMarkerPrefix.ReplaceAllString(block, ""))
}

// findInboxBlockByHash returns (blockIndex, block, true) for the first
// non-header block whose hash matches msgHash. Returns (-1, "", false) if no
// match is found.
func findInboxBlockByHash(content, msgHash string) (int, string, bool) {
	blocks := readBlocks(content)
	for i, block := range blocks {
		if inboxHeaderRegex.MatchString(block) {
			continue
		}
		if inboxBlockHash(block) == msgHash {
			return i, block, true
		}
	}
	return -1, "", false
}

// appendToInbox writes a new entry to Inbox.md and returns its stable hash.
func (b *Bot) appendToInbox(content string, timezone *time.Location) (string, error) {
	exists, err := b.fs.Exists(fs.DirUserRoot, fs.InboxFilename)
	if err != nil {
		return "", fmt.Errorf("appendToInbox: %w", err)
	}

	content = strings.TrimSpace(content)

	var md string
	if exists {
		md, err = b.fs.Read(fs.DirUserRoot, fs.InboxFilename)
		if err != nil {
			return "", fmt.Errorf("appendToInbox: %w", err)
		}
		md = txt.NormNewLines(md)
		md = strings.TrimSpace(md)
		if len(md) != 0 {
			md += "\n"
		}
	}

	// Add today's header if it doesn't exist
	if !strings.Contains(md, todayHeader(timezone)) {
		md += todayHeader(timezone) + "\n"
	}

	// Format timestamp with timezone
	// TODO should we use timezone here?
	timestamp := now().In(timezone).Format("`15:04`")

	newEntry := fmt.Sprintf("- [ ] %s %s", timestamp, content)
	md += newEntry + "\n"

	if err := b.fs.Write(fs.DirUserRoot, fs.InboxFilename, md); err != nil {
		return "", fmt.Errorf("appendToInbox: %w", err)
	}

	return inboxBlockHash(newEntry), nil
}

// moveFromInbox passes the messages identified by msgHashes to the callback.
// On callback success, it removes those messages from the chat file.
// A msgHash is the stable hash returned by inboxBlockHash; it survives the
// `[ ]` ↔ `[x]` completion toggle.
// On collapse=false the callback is called once per message.
func (b *Bot) moveFromInbox(
	callback func(content string, timestamp time.Time) error,
	collapse bool,
	msgHashes ...string,
) error {
	key, err := b.fs.SafePath(fs.DirUserRoot, "")
	if err != nil {
		return fmt.Errorf("failed to get safe path: %w", err)
	}

	lock := userLock(key)
	lock.Lock()
	defer lock.Unlock()

	content, err := b.fs.Read(fs.DirUserRoot, fs.InboxFilename)
	if err != nil {
		return err
	}

	blocks := readBlocks(content)

	// Build hash -> block-index for every non-header block. Validate that all
	// requested hashes resolve to real blocks.
	hashToBlockIndex := make(map[string]int)
	hasAnyMsg := false
	for i, block := range blocks {
		if inboxHeaderRegex.MatchString(block) {
			continue
		}
		hasAnyMsg = true
		hashToBlockIndex[inboxBlockHash(block)] = i
	}
	if !hasAnyMsg {
		return fmt.Errorf("no messages found")
	}
	resolvedBlockIndices := make([]int, 0, len(msgHashes))
	for _, h := range msgHashes {
		idx, ok := hashToBlockIndex[h]
		if !ok {
			return fmt.Errorf("msgHash %q not found in inbox", h)
		}
		resolvedBlockIndices = append(resolvedBlockIndices, idx)
	}

	// Process in ascending block-index order so removal later is deterministic.
	sort.Ints(resolvedBlockIndices)

	// Collect specified messages from inbox.
	var msgs []struct {
		content   string
		timestamp time.Time
		index     int
	}
	for _, blockIndex := range resolvedBlockIndices {
		block := blocks[blockIndex]

		// Find closest header above target msg for date context
		var headerDate string
		for i := blockIndex - 1; i >= 0; i-- {
			if inboxHeaderRegex.MatchString(blocks[i]) {
				headerDate = blocks[i]
				break
			}
		}

		// Extract time and get full content. Tolerate optional Markdown-task
		// prefix `- [ ] ` / `- [x] ` (new inbox format); legacy entries without
		// the prefix also match.
		timestampRegex := regexp.MustCompile(`^(?:- \[[ xX]\] )?` + "`" + `(\d{2}:\d{2})` + "`" + ` `)
		timeMatch := timestampRegex.FindStringSubmatch(block)
		if len(timeMatch) < 2 {
			return fmt.Errorf("failed to parse msg timestamp for block %d", blockIndex)
		}

		timeStr := timeMatch[1]
		// Remove the full matched prefix (optional checkbox + timestamp + space).
		recordContent := block[len(timeMatch[0]):]

		// Parse full timestamp from header date + time
		dateRegex := regexp.MustCompile(`^#### (\d{1,2}) ([A-Za-z]+), [A-Za-z]+`)
		dateMatches := dateRegex.FindStringSubmatch(headerDate)
		if len(dateMatches) < 3 {
			return fmt.Errorf("failed to parse header date for block %d", blockIndex)
		}

		// Build full timestamp
		dateTimeStr := fmt.Sprintf("%s %s %s", dateMatches[1], dateMatches[2], timeStr)
		timestamp, err := time.Parse("2 January 15:04", dateTimeStr)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp for block %d: %w", blockIndex, err)
		}

		msgs = append(msgs, struct {
			content   string
			timestamp time.Time
			index     int
		}{
			content:   recordContent,
			timestamp: timestamp,
			index:     blockIndex,
		})
	}

	// First we save all the messages to files, only then we remove them from the inbox.
	if collapse {
		content := strings.Builder{}
		for _, msg := range msgs {
			content.WriteString(msg.content)
			content.WriteString("\n")
		}
		err = callback(strings.TrimSpace(content.String()), msgs[0].timestamp)
		if err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}
	} else {
		for _, msg := range msgs {
			if err := callback(msg.content, msg.timestamp); err != nil {
				return fmt.Errorf("callback failed: %w", err)
			}
		}
	}

	blocksToRemove := make(map[int]bool)
	for _, msg := range msgs {
		blocksToRemove[msg.index] = true
	}
	newBlocks := make([]string, 0)
	for i, block := range blocks {
		if blocksToRemove[i] {
			continue
		}
		newBlocks = append(newBlocks, block)
	}
	modifiedContent := strings.TrimSpace(strings.Join(newBlocks, "\n"))

	return b.fs.Write(fs.DirUserRoot, fs.InboxFilename, modifiedContent)
}

// readBlocks parses content into logical blocks
// Returns slice where each element is either a header or a complete record
func readBlocks(content string) []string {
	content = txt.NormNewLines(content)
	lines := strings.Split(content, "\n")

	headerRegex := regexp.MustCompile(`^#### `)
	timestampRegex := regexp.MustCompile(`^(?:- \[[ xX]\] )?` + "`" + `\d{2}:\d{2}` + "`" + ` `)

	var blocks []string
	var currentBlock strings.Builder

	for _, line := range lines {
		isHeader := headerRegex.MatchString(line)
		isTimestamp := timestampRegex.MatchString(line)

		if isHeader {
			// Save previous block if exists
			if currentBlock.Len() > 0 {
				blocks = append(blocks, strings.TrimSpace(currentBlock.String()))
				currentBlock.Reset()
			}
			// DisplayName is always its own block
			blocks = append(blocks, line)
		} else if isTimestamp {
			// Save previous block if exists
			if currentBlock.Len() > 0 {
				blocks = append(blocks, strings.TrimSpace(currentBlock.String()))
				currentBlock.Reset()
			}
			// Start new block with timestamp
			currentBlock.WriteString(line)
		} else {
			// Continue current block or start new block
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
				currentBlock.WriteString(line)
			} else {
				currentBlock.WriteString(line)
			}
		}
	}

	// Add final block
	if currentBlock.Len() > 0 {
		blocks = append(blocks, strings.TrimSpace(currentBlock.String()))
	}

	return blocks
}

func todayHeader(timezone *time.Location) string {
	nowTZ := now().In(timezone)
	return fmt.Sprintf("#### %d %s, %s", nowTZ.Day(), nowTZ.Format("January"), nowTZ.Weekday())
}

func userLock(rootPath string) *sync.Mutex {
	mu.Lock()
	defer mu.Unlock()

	if userLocks == nil {
		userLocks = make(map[string]*sync.Mutex)
	}
	if lock, exists := userLocks[rootPath]; exists {
		return lock
	}

	newLock := &sync.Mutex{}
	userLocks[rootPath] = newLock

	return newLock
}
