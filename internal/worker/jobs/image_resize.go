package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type imageResizePayload struct {
	SourceURL string `json:"source_url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Quality   int    `json:"quality"`
}

func ImageResizeHandler(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p imageResizePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if p.SourceURL == "" {
		return nil, fmt.Errorf("source_url is required")
	}
	if p.Width == 0 || p.Height == 0 {
		return nil, fmt.Errorf("width and height are required")
	}

	sleepDuration := 200*time.Millisecond + time.Duration(rand.Intn(600))*time.Millisecond
	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if rand.Float64() < 0.05 {
		return nil, fmt.Errorf("simulated upstream failure")
	}

	format := p.Format
	if format == "" {
		format = "webp"
	}

	result := map[string]string{
		"output_url": fmt.Sprintf("https://cdn.example.com/resized-%d.%s", rand.Int63(), format),
	}
	return json.Marshal(result)
}
