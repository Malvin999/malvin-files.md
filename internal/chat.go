package internal

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/pkg/txt"
)

var (
	Now       = time.Now
	mu        sync.Mutex
	userLocks map[string]*sync.Mutex
)

func (b *Bot) saveToChat(content string, timezone *time.Location) (int, error) {
	exists, err := b.fs.Exists(fs.DirRoot, fs.ChatFilename)
	if err != nil {
		return 0, fmt.Errorf("saveToChat: %w", err)
	}

	content = strings.TrimSpace(content)

	var md string
	if exists {
		md, err = b.fs.Read(fs.DirRoot, fs.ChatFilename)
		if err != nil {
			return 0, fmt.Errorf("saveToChat: %w", err)
		}
		md = txt.NormNewLines(md)
		md = strings.TrimSpace(md)
		if len(md) != 0 {
			md += "\n"
		}
	}

	// Count existing records before adding new one
	blocks := readMessages(md)
	headerRegex := regexp.MustCompile(`^#### `)
	recordCount := 0
	for _, block := range blocks {
		if !headerRegex.MatchString(block) {
			recordCount++
		}
	}

	// Add today's header if it doesn't exist
	if !strings.Contains(md, todayHeader(timezone)) {
		md += todayHeader(timezone) + "\n"
	}

	// Format timestamp with timezone
	timestamp := time.Now().In(timezone).Format("`15:04`")

	// Handle images similar to journal
	if txt.HasImage(content) {
		// If there's an image - place timestamp under the image
		re := regexp.MustCompile(txt.ImgPattern)
		imgLink := re.FindString(content)
		content = strings.TrimSpace(strings.Replace(content, imgLink, "", 1))
		content = fmt.Sprintf("%s\n%s %s\n", imgLink, timestamp, strings.TrimSpace(content))
	} else {
		content = fmt.Sprintf("%s %s\n", timestamp, content)
	}

	md += content

	if err := b.fs.Write(fs.DirRoot, fs.ChatFilename, md); err != nil {
		return 0, fmt.Errorf("saveToChat: %w", err)
	}

	// Return index of the newly added record (1-based)
	return recordCount + 1, nil
}

func (b *Bot) MoveRecordFromChat(callback func(content string, timestamp time.Time) error, indices ...int) error {
	key, err := b.fs.SafePath(fs.DirRoot, "")
	if err != nil {
		return fmt.Errorf("failed to get safe path: %w", err)
	}

	lock := userLock(key)
	lock.Lock()
	defer lock.Unlock()

	content, err := b.fs.Read(fs.DirRoot, fs.ChatFilename)
	if err != nil {
		return err
	}

	blocks := readMessages(content)

	// Filter to find record blocks (not headers)
	headerRegex := regexp.MustCompile(`^#### `)
	var recordIndices []int

	for i, block := range blocks {
		if !headerRegex.MatchString(block) {
			recordIndices = append(recordIndices, i)
		}
	}

	if len(recordIndices) == 0 {
		return fmt.Errorf("no records found")
	}

	// Validate all indices
	for _, index := range indices {
		if index < 0 || index >= len(recordIndices) {
			return fmt.Errorf("index %d out of bounds: use 0-%d", index, len(recordIndices)-1)
		}
	}

	// Sort indices in descending order to avoid index shifting issues when removing
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	// Remove duplicates
	uniqueIndices := make([]int, 0, len(indices))
	seen := make(map[int]bool)
	for _, index := range indices {
		if !seen[index] {
			uniqueIndices = append(uniqueIndices, index)
			seen[index] = true
		}
	}

	// Track which block indices to remove
	blocksToRemove := make(map[int]bool)

	// Process each record
	for _, index := range uniqueIndices {
		targetBlockIndex := recordIndices[index]
		targetRecord := blocks[targetBlockIndex]

		// Find closest header above target record for date context
		var headerDate string
		for i := targetBlockIndex - 1; i >= 0; i-- {
			if headerRegex.MatchString(blocks[i]) {
				headerDate = blocks[i]
				break
			}
		}

		// Extract time from record and content without timestamp
		timestampRegex := regexp.MustCompile(`^` + "`" + `(\d{2}:\d{2})` + "`" + ` (.*)`)
		matches := timestampRegex.FindStringSubmatch(targetRecord)
		if len(matches) < 3 {
			return fmt.Errorf("failed to parse record timestamp for index %d", index)
		}

		timeStr := matches[1]
		recordContent := matches[2]

		// Parse full timestamp from header date + time
		dateRegex := regexp.MustCompile(`^#### (\d{1,2}) ([A-Za-z]+), [A-Za-z]+`)
		dateMatches := dateRegex.FindStringSubmatch(headerDate)
		if len(dateMatches) < 3 {
			return fmt.Errorf("failed to parse header date for index %d", index)
		}

		// Build full timestamp
		dateTimeStr := fmt.Sprintf("%s %s %s", dateMatches[1], dateMatches[2], timeStr)
		timestamp, err := time.Parse("2 January 15:04", dateTimeStr)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp for index %d: %w", index, err)
		}

		// Call callback with content and timestamp
		if err := callback(recordContent, timestamp); err != nil {
			return fmt.Errorf("callback failed for index %d: %w", index, err)
		}

		// Mark block for removal
		blocksToRemove[targetBlockIndex] = true
	}

	// Remove target blocks and rebuild content
	newBlocks := make([]string, 0, len(blocks)-len(blocksToRemove))
	for i, block := range blocks {
		if !blocksToRemove[i] {
			newBlocks = append(newBlocks, block)
		}
	}

	modifiedContent := strings.TrimSpace(strings.Join(newBlocks, "\n"))

	return b.fs.Write(fs.DirRoot, fs.ChatFilename, modifiedContent)
}

// readMessages parses content into logical blocks
// Returns slice where each element is either a header or a complete record
func readMessages(content string) []string {
	content = txt.NormNewLines(content)
	lines := strings.Split(content, "\n")

	headerRegex := regexp.MustCompile(`^#### `)
	timestampRegex := regexp.MustCompile(`^` + "`" + `\d{2}:\d{2}` + "`" + ` `)

	var blocks []string
	var currentBlock strings.Builder

	for _, line := range lines {
		isHeader := headerRegex.MatchString(line)
		isTimestamp := timestampRegex.MatchString(line)

		if isHeader || isTimestamp {
			// Save previous block if exists
			if currentBlock.Len() > 0 {
				blocks = append(blocks, strings.TrimSpace(currentBlock.String()))
				currentBlock.Reset()
			}

			// Start new block
			currentBlock.WriteString(line)
		} else {
			// Continue current block
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
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
	nowTZ := time.Now().In(timezone)
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
