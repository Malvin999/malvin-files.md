package server

import "strings"

func Merge(original, modified string) string {
	// If one string is a prefix of the other, just return the longer one
	if strings.HasPrefix(modified, original) {
		return modified
	}
	if strings.HasPrefix(original, modified) {
		return original
	}

	// Otherwise, perform a line-by-line merge
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	// Find the common prefix of lines
	var commonPrefixLength int
	for commonPrefixLength < len(originalLines) &&
		commonPrefixLength < len(modifiedLines) &&
		originalLines[commonPrefixLength] == modifiedLines[commonPrefixLength] {
		commonPrefixLength++
	}

	// Create a set of lines from both files after the common prefix
	uniqueLines := make(map[string]bool)
	for i := commonPrefixLength; i < len(originalLines); i++ {
		if originalLines[i] != "" {
			uniqueLines[originalLines[i]] = true
		}
	}
	for i := commonPrefixLength; i < len(modifiedLines); i++ {
		if modifiedLines[i] != "" {
			uniqueLines[modifiedLines[i]] = true
		}
	}

	// Build the result: common prefix + all unique lines
	result := strings.Join(originalLines[:commonPrefixLength], "\n")

	// Add the unique lines
	for line := range uniqueLines {
		result += "\n" + line
	}

	return result
}
