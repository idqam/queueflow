package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

func AlwaysFailHandler(_ context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Reason    string `json:"reason"`
		ErrorCode int    `json:"error_code"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if p.ErrorCode == 0 {
		p.ErrorCode = 500
	}

	slog.Info("always_fail handler invoked", "reason", p.Reason, "error_code", p.ErrorCode)
	return nil, fmt.Errorf("simulated failure: %s (code %d)", p.Reason, p.ErrorCode)
}
