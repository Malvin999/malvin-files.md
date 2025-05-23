package server

import "strings"

// Merge combines two strings (s1 and s2) by identifying longest sequences of common lines.
//
// The algorithm:
// 1) Splits both inputs into lines
// 2) Uses dynamic programming to find the longest common subsequence (LCS) between the lines
// 3) Constructs a merged result that preserves all unique content from both strings
// 4) Maintains the original order of content from both strings
func Merge(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	if len(s2) == 0 {
		return s1
	}
	lines1 := strings.Split(s1, "\n")
	lines2 := strings.Split(s2, "\n")

	// Dynamical programming table containing the longest common prefix for each pair.
	lcsLength := make([][]int, len(lines1)+1)
	for i := range lcsLength {
		lcsLength[i] = make([]int, len(lines2)+1)
	}

	// Fill the lcsLength table.
	for i := 1; i <= len(lines1); i++ {
		for j := 1; j <= len(lines2); j++ {
			if lines1[i-1] == lines2[j-1] {
				lcsLength[i][j] = lcsLength[i-1][j-1] + 1
			} else {
				lcsLength[i][j] = max(lcsLength[i-1][j], lcsLength[i][j-1])
			}
		}
	}

	// Build the merged result.
	result := backtrack(lines1, lines2, lcsLength, len(lines1), len(lines2))
	result = mergeHeaders(result)

	return strings.Join(result, "\n")
}

// backtrack performs backtracking through the dynamic programming table lcsLength
// to construct the merged result based on the longest common subsequence (LCS).
func backtrack(lines1, lines2 []string, lcsLength [][]int, i, j int) []string {
	if i == 0 && j == 0 {
		return []string{}
	}

	if i == 0 {
		return append(backtrack(lines1, lines2, lcsLength, i, j-1), lines2[j-1])
	}

	if j == 0 {
		return append(backtrack(lines1, lines2, lcsLength, i-1, j), lines1[i-1])
	}

	// If the current lines are the same, include it only once.
	if lines1[i-1] == lines2[j-1] {
		return append(backtrack(lines1, lines2, lcsLength, i-1, j-1), lines1[i-1])
	}

	// Choose the direction with the longer common subsequence.
	if lcsLength[i-1][j] > lcsLength[i][j-1] {
		return append(backtrack(lines1, lines2, lcsLength, i-1, j), lines1[i-1])
	} else {
		return append(backtrack(lines1, lines2, lcsLength, i, j-1), lines2[j-1])
	}
}

// Headers like this should be merged:
// #### 23 May, Friday 🤸‍🍽💪💧
// #### 23 May, Friday 🤸‍🍽💪
// #### 23 May, Friday 🤸‍🍽
func mergeHeaders(lines []string) []string {
	var result []string

	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "####") {
			result = append(result, lines[i])
			continue
		}

		// Find where emojis start
		split := strings.LastIndex(lines[i], " ")
		base := lines[i][:split+1]
		chars := lines[i][split+1:]

		// Collect chars from similar headers
		for j := i + 1; j < len(lines) && strings.HasPrefix(lines[j], base); j++ {
			chars += lines[j][split+1:]
			i = j
		}

		// Unique chars
		seen := map[rune]bool{}
		var unique strings.Builder
		for _, r := range chars {
			if !seen[r] {
				seen[r] = true
				unique.WriteRune(r)
			}
		}

		result = append(result, base+unique.String())
	}

	return result
}
