package harness

import (
	"fmt"
	"os"
	"strings"
)

type LocalMemoryFallback struct{}

func NewLocalMemoryFallback() *LocalMemoryFallback {
	return &LocalMemoryFallback{}
}

func (l *LocalMemoryFallback) AdapterName() string {
	return FallbackMode
}

func (l *LocalMemoryFallback) Save(event *MemoryEvent) error {
	entry := l.formatEntry(event)

	if err := ensureDecisionsHeader(); err != nil {
		return err
	}

	f, err := os.OpenFile(DecisionsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", DecisionsFile, err)
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func (l *LocalMemoryFallback) formatEntry(event *MemoryEvent) string {
	var sb strings.Builder

	sb.WriteString("\n---\n\n")
	sb.WriteString(fmt.Sprintf("## [%s] %s — %s\n", event.Timestamp, event.Type, event.Title))
	sb.WriteString(fmt.Sprintf("**topic_key**: %s\n", event.TopicKey))
	sb.WriteString(fmt.Sprintf("**saved_via**: fallback (Engram not available)\n\n"))
	sb.WriteString(event.Content)
	sb.WriteString("\n")

	return sb.String()
}

func ensureDecisionsHeader() error {
	info, err := os.Stat(DecisionsFile)
	if err == nil && info.Size() > 0 {
		return nil
	}

	header := `# Decisions Log

This file is the LOCAL FALLBACK for memory events when Engram is not available.
Events are formatted with the standard **What / Why / Where / Learned** structure.

When Engram is enabled, events go to ` + "`.harness/memory-queue.json`" + ` for sync instead.
`
	return WriteFile(DecisionsFile, header)
}

func CountLocalEntries() int {
	data, err := os.ReadFile(DecisionsFile)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n---\n")
}
