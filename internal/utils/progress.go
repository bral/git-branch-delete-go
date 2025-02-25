package utils

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// Progress represents a progress indicator
type Progress struct {
	spinner *spinner.Spinner
	message string
	writer  io.Writer
}

// NewProgress creates a new progress indicator
func NewProgress(message string) *Progress {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = os.Stderr

	return &Progress{
		spinner: s,
		message: message,
		writer:  os.Stderr,
	}
}

// Start begins showing the progress indicator
func (p *Progress) Start() {
	p.spinner.Suffix = fmt.Sprintf(" %s", p.message)
	p.spinner.Start()
}

// Stop ends the progress indicator
func (p *Progress) Stop() {
	p.spinner.Stop()
}

// Update changes the progress message
func (p *Progress) Update(message string) {
	p.message = message
	p.spinner.Suffix = fmt.Sprintf(" %s", message)
}

// Success stops the spinner and shows a success message
func (p *Progress) Success(message string) {
	p.Stop()
	fmt.Fprintf(p.writer, "✓ %s\n", message)
}

// Error stops the spinner and shows an error message
func (p *Progress) Error(message string) {
	p.Stop()
	fmt.Fprintf(p.writer, "✗ %s\n", message)
}

// MultiProgress handles multiple progress indicators
type MultiProgress struct {
	items    []*Progress
	messages []string
}

// NewMultiProgress creates a new multi-progress indicator
func NewMultiProgress() *MultiProgress {
	return &MultiProgress{}
}

// Add adds a new progress item
func (mp *MultiProgress) Add(message string) *Progress {
	p := NewProgress(message)
	mp.items = append(mp.items, p)
	mp.messages = append(mp.messages, message)
	return p
}

// Start begins showing all progress indicators
func (mp *MultiProgress) Start() {
	for _, p := range mp.items {
		p.Start()
	}
}

// Stop ends all progress indicators
func (mp *MultiProgress) Stop() {
	for _, p := range mp.items {
		p.Stop()
	}
}
