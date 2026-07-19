package harness

import (
	"encoding/json"
	"fmt"
	"os"
)

const EventPending = "pending"
const EventSynced = "synced"

type EngramMemoryAdapter struct{}

func NewEngramMemoryAdapter() *EngramMemoryAdapter {
	return &EngramMemoryAdapter{}
}

func (e *EngramMemoryAdapter) AdapterName() string {
	return EngramMode
}

func (e *EngramMemoryAdapter) Save(event *MemoryEvent) error {
	queue, err := e.LoadQueue()
	if err != nil {
		return fmt.Errorf("engram adapter: cannot load queue: %w", err)
	}

	event.Status = EventPending
	queue = append(queue, *event)

	return e.WriteQueue(queue)
}

func (e *EngramMemoryAdapter) LoadQueue() ([]MemoryEvent, error) {
	data, err := os.ReadFile(MemoryQueueFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []MemoryEvent{}, nil
		}
		return nil, err
	}

	var queue []MemoryEvent
	if err := json.Unmarshal(data, &queue); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", MemoryQueueFile, err)
	}
	return queue, nil
}

func (e *EngramMemoryAdapter) WriteQueue(queue []MemoryEvent) error {
	data, err := json.MarshalIndent(queue, "", "  ")
	if err != nil {
		return err
	}
	return WriteFile(MemoryQueueFile, string(data))
}

func (e *EngramMemoryAdapter) PendingEvents() []MemoryEvent {
	queue, err := e.LoadQueue()
	if err != nil {
		return nil
	}
	var pending []MemoryEvent
	for _, ev := range queue {
		if ev.Status == EventPending {
			pending = append(pending, ev)
		}
	}
	return pending
}

func (e *EngramMemoryAdapter) MarkAllSynced() error {
	queue, err := e.LoadQueue()
	if err != nil {
		return err
	}
	for i := range queue {
		if queue[i].Status == EventPending {
			queue[i].Status = EventSynced
		}
	}
	return e.WriteQueue(queue)
}

func (e *EngramMemoryAdapter) Stats() (total int, pending int, synced int) {
	queue, err := e.LoadQueue()
	if err != nil {
		return 0, 0, 0
	}
	total = len(queue)
	for _, ev := range queue {
		if ev.Status == EventPending {
			pending++
		} else if ev.Status == EventSynced {
			synced++
		}
	}
	return total, pending, synced
}
