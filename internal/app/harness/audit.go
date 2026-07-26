package harness

import (
	"encoding/json"
	"fmt"
)

const TransitionAuditFile = ".harness/transition-audit.jsonl"

type TransitionAuditEvent struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	From      string `json:"from"`
	To        string `json:"to"`
	Gate      string `json:"gate,omitempty"`
	Result    string `json:"result"`
	Reason    string `json:"reason,omitempty"`
}

func AppendTransitionAudit(event TransitionAuditEvent) error {
	if event.Timestamp == "" {
		event.Timestamp = NowISO()
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cannot marshal transition audit event: %w", err)
	}
	return AppendFile(TransitionAuditFile, string(data)+"\n")
}
