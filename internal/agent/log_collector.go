package agent

import (
	"sync"

	"v2wall/internal/logwriter"
)

type LogCollector struct {
	mu    sync.Mutex
	items []logwriter.LogEntry
}

func NewLogCollector() *LogCollector {
	return &LogCollector{}
}

func (c *LogCollector) Add(entry logwriter.LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = append(c.items, entry)
}

func (c *LogCollector) Flush() []logwriter.LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) == 0 {
		return nil
	}
	ret := c.items
	c.items = nil
	return ret
}
