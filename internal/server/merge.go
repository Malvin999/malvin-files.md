package server

import "strings"

// Merge combines two strings (s1 and s2) by identifying longest common sections
// and unique content. This function is particularly useful for merging text that may have
// been edited independently, such as journal entries or notes.
//
// The algorithm:
// - Splits both inputs into lines
// - Uses dynamic programming to find the longest common subsequence (LCS) between the lines
// - Constructs a merged result that preserves all unique content from both strings
// - Maintains the original order of content from both strings.
func Merge(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	if len(s2) == 0 {
		return s1
	}
	lines1 := strings.Split(s1, "\n")
	lines2 := strings.Split(s2, "\n")

	lcsLengths := make([][]int, len(lines1)+1)
	for i := range lcsLengths {
		lcsLengths[i] = make([]int, len(lines2)+1)
	}

	// Fill the lcsLengths table.
	for i := 1; i <= len(lines1); i++ {
		for j := 1; j <= len(lines2); j++ {
			if lines1[i-1] == lines2[j-1] {
				lcsLengths[i][j] = lcsLengths[i-1][j-1] + 1
			} else {
				lcsLengths[i][j] = max(lcsLengths[i-1][j], lcsLengths[i][j-1])
			}
		}
	}

	// Build the merged result.
	result := backgrack(lines1, lines2, lcsLengths, len(lines1), len(lines2))
	return strings.Join(result, "\n")
}

// backtrack performs backtracking through the dynamic programming table lcsLengths
// to construct the merged result based on the longest common subsequence (LCS).
func backgrack(lines1, lines2 []string, lcsLengths [][]int, i, j int) []string {
	if i == 0 && j == 0 {
		return []string{}
	}

	if i == 0 {
		return append(backgrack(lines1, lines2, lcsLengths, i, j-1), lines2[j-1])
	}

	if j == 0 {
		return append(backgrack(lines1, lines2, lcsLengths, i-1, j), lines1[i-1])
	}

	// If the current lines are the same, include it only once.
	if lines1[i-1] == lines2[j-1] {
		return append(backgrack(lines1, lines2, lcsLengths, i-1, j-1), lines1[i-1])
	}

	// Choose the direction with the longer common subsequence.
	if lcsLengths[i-1][j] > lcsLengths[i][j-1] {
		return append(backgrack(lines1, lines2, lcsLengths, i-1, j), lines1[i-1])
	} else {
		return append(backgrack(lines1, lines2, lcsLengths, i, j-1), lines2[j-1])
	}
}
