package ui

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	progressTemplate = `{{ .Prefix }} {{ .Progress }} of {{ .Total }} branches processed`
	progressBarWidth = 40
)

// KeyboardShortcuts defines available keyboard shortcuts
var KeyboardShortcuts = map[rune]string{
	'q': "quit",
	'a': "select all",
	'n': "select none",
	'/': "search",
	'?': "help",
}

type ProgressBar struct {
	mu      sync.Mutex
	current int
	total   int
	prefix  string
	start   time.Time
}

func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		prefix: prefix,
		start:  time.Now(),
	}
}

func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
	p.render()
}

func (p *ProgressBar) render() {
	if p.total == 0 {
		return
	}

	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(progressBarWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", progressBarWidth-filled)

	elapsed := time.Since(p.start)
	eta := time.Duration(float64(elapsed) / percent * (1 - percent))

	status := fmt.Sprintf("%s [%s] %d/%d (%d%%) ETA: %s",
		p.prefix,
		bar,
		p.current,
		p.total,
		int(percent*100),
		formatDuration(eta),
	)

	// Clear line and render progress
	fmt.Printf("\r\033[K%s", status)
}

func (p *ProgressBar) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = p.total
	p.render()
	fmt.Println() // Move to next line when done
}

type FuzzySearch struct {
	query         string
	maxDistance   int
	caseSensitive bool
}

func NewFuzzySearch(query string, maxDistance int, caseSensitive bool) *FuzzySearch {
	return &FuzzySearch{
		query:         query,
		maxDistance:   maxDistance,
		caseSensitive: caseSensitive,
	}
}

func (fs *FuzzySearch) Match(text string) bool {
	if !fs.caseSensitive {
		text = strings.ToLower(text)
		fs.query = strings.ToLower(fs.query)
	}

	// Simple fuzzy matching using Levenshtein distance
	distance := levenshteinDistance(fs.query, text)
	return distance <= fs.maxDistance
}

func levenshteinDistance(s1, s2 string) int {
	if s1 == "" {
		return len(s2)
	}
	if s2 == "" {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = minInt(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[len(s1)][len(s2)]
}

func minInt(numbers ...int) int {
	if len(numbers) == 0 {
		return 0
	}
	result := numbers[0]
	for _, num := range numbers[1:] {
		result = int(math.Min(float64(result), float64(num)))
	}
	return result
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func ShowHelp() {
	fmt.Println("\nKeyboard shortcuts:")
	for key, action := range KeyboardShortcuts {
		fmt.Printf("  %s: %s\n", color.CyanString(string(key)), action)
	}
	fmt.Println()
}
