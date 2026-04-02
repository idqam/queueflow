package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func DataTransformHandler(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Source            string   `json:"source"`
		Operations        []string `json:"operations"`
		Output            string   `json:"output"`
		RowCountEstimate  int      `json:"row_count_estimate"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	sleepDuration := 5*time.Second + time.Duration(rand.Intn(10000))*time.Millisecond
	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	rowCount := p.RowCountEstimate
	if rowCount == 0 {
		rowCount = 50000
	}

	result := map[string]any{
		"rows_processed": rowCount,
		"output":         p.Output,
	}
	return json.Marshal(result)
}
